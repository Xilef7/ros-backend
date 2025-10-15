package test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"strconv"
	"testing"
	"time"

	"restaurant-ordering-system/api/proto"
	"restaurant-ordering-system/internal/pkg/auth"
	"restaurant-ordering-system/internal/pkg/config"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestEndToEndFlow(t *testing.T) {
	// 1. Load configuration
	cfgPath := "../configs/config.test.json"
	cfg, err := config.LoadConfig(cfgPath)
	require.NoError(t, err, "failed to load config")

	// 2. Generate TLS certificate
	certPEM, keyPEM, err := GenerateSelfSignedTLSCert()
	require.NoError(t, err, "failed to generate TLS certificate")

	// 3. Create network
	net, err := network.New(t.Context())
	testcontainers.CleanupNetwork(t, net)
	require.NoError(t, err, "failed to create network")

	// 4. Start PostgreSQL container
	pgC, err := postgres.Run(t.Context(), "postgres:17-alpine",
		postgres.WithDatabase(cfg.Database.Database),
		postgres.WithUsername(cfg.Database.User),
		postgres.WithPassword(cfg.Database.Password),
		postgres.WithInitScripts("../migrations/001_create_tables.sql"),
		postgres.WithSQLDriver("pgx"),
		postgres.BasicWaitStrategies(),
		network.WithNetwork([]string{cfg.Database.Host}, net),
	)
	testcontainers.CleanupContainer(t, pgC)
	require.NoError(t, err, "failed to start postgres container")

	// 5. Start Redis container
	rC, err := redis.Run(context.Background(), "redis:8",
		network.WithNetwork([]string{cfg.Redis.Host}, net),
	)
	testcontainers.CleanupContainer(t, rC)
	require.NoError(t, err, "failed to start redis container")

	// 6. Start gRPC server in docker container
	svrPort, err := nat.NewPort("tcp", strconv.Itoa(cfg.Server.Port))
	require.NoError(t, err, "failed to create port instance")

	svrC, err := testcontainers.Run(t.Context(), "",
		testcontainers.WithDockerfile(testcontainers.FromDockerfile{
			Context: "..",
		}),
		testcontainers.WithFiles(
			testcontainers.ContainerFile{
				HostFilePath:      cfgPath,
				ContainerFilePath: "/app/configs/config.json",
			},
			testcontainers.ContainerFile{
				Reader:            bytes.NewReader(certPEM),
				ContainerFilePath: "/app/certs/server_cert.pem",
			},
			testcontainers.ContainerFile{
				Reader:            bytes.NewReader(keyPEM),
				ContainerFilePath: "/app/certs/server_key.pem",
			},
		),
		testcontainers.WithExposedPorts(svrPort.Port()),
		testcontainers.WithWaitStrategy(wait.ForListeningPort(svrPort)),
		network.WithNetwork(nil, net),
	)
	testcontainers.CleanupContainer(t, svrC)
	require.NoError(t, err, "failed to start server container")

	// 7. Create gRPC client
	svrHost, err := svrC.Host(t.Context())
	require.NoError(t, err, "failed to get server host")

	svrMappedPort, err := svrC.MappedPort(t.Context(), svrPort)
	require.NoError(t, err, "failed to get server port")

	cp := x509.NewCertPool()
	cp.AppendCertsFromPEM(certPEM)
	tlsCreds := credentials.NewClientTLSFromCert(cp, "localhost")
	conn, err := grpc.NewClient(svrHost+":"+svrMappedPort.Port(), grpc.WithTransportCredentials(tlsCreds))
	require.NoError(t, err, "failed to create gRPC channel")
	defer conn.Close()

	// 8. Create gRPC clients
	customerClient := proto.NewCustomerServiceClient(conn)
	authClient := proto.NewAuthServiceClient(conn)
	menuClient := proto.NewMenuServiceClient(conn)
	orderClient := proto.NewOrderServiceClient(conn)
	tabClient := proto.NewTabServiceClient(conn)

	// 9. Get admin token
	adminToken, err := auth.GenerateAdminJWT([]byte(cfg.JWT.Secret), time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, adminToken)
	adminCred := grpc.PerRPCCredentials(oauth.TokenSource{
		TokenSource: oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: adminToken,
		}),
	})

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	// 10. Initialize menu
	createMenuReq := &proto.CreateMenuItemRequest{}
	menuItem := &proto.MenuItem{}
	menuItem.SetName("Test Menu")
	menuItem.SetPrice(1)
	menuItem.SetPortionSize(1)
	menuItem.SetAvailable(true)
	createMenuReq.SetMenuItem(menuItem)
	createMenuResp, err := menuClient.CreateMenuItem(ctx, createMenuReq, adminCred)
	require.NoError(t, err)
	require.NotEmpty(t, createMenuResp.GetId())

	// 11. Create customer
	loginID := "testcustomer"
	password := "testcustomer"
	custReq := &proto.CreateCustomerRequest{}
	custReq.SetLoginId(loginID)
	custReq.SetEmail("test.customer@gmail.com")
	custReq.SetPassword(password)
	custReq.SetName("Test Customer")
	custReq.SetPhoneNumber("0123456789")
	cust, err := customerClient.CreateCustomer(ctx, custReq)
	require.NoError(t, err)
	require.NotEmpty(t, cust.GetName())

	// 12. Simulate E2E flow
	// a. Login as customer
	genTokenReq := &proto.GenerateTokenRequest{}
	genTokenReq.SetLoginId(loginID)
	genTokenReq.SetPassword(password)
	token, err := authClient.GenerateToken(ctx, genTokenReq)
	require.NoError(t, err)
	accessToken := token.GetAccessToken()
	require.NotEmpty(t, accessToken)
	customerCred := grpc.PerRPCCredentials(oauth.TokenSource{
		TokenSource: oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: accessToken,
		}),
	})

	// b. Create tab
	tabResp, err := tabClient.CreateTab(ctx, &emptypb.Empty{}, adminCred)
	require.NoError(t, err)
	require.NotEmpty(t, tabResp.GetId())

	// c. Visit tab
	visitTabReq := &proto.VisitTabRequest{}
	visitTabReq.SetTabId(tabResp.GetId())
	visitTabReq.SetCustomerId(cust.GetId())
	_, err = tabClient.VisitTab(ctx, visitTabReq, customerCred)
	require.NoError(t, err)

	// d. Get tab
	getTabReq := &proto.GetOpenTabRequest{}
	getTabReq.SetTabId(tabResp.GetId())
	openTab, err := tabClient.GetOpenTab(ctx, getTabReq)
	require.NoError(t, err)
	require.NotEmpty(t, openTab.GetId())
	order := openTab.GetOrders()[0]

	// e. List menu items
	menu, err := menuClient.ListMenuItems(ctx, &emptypb.Empty{})
	require.NoError(t, err)
	items := menu.GetItems()
	require.NotEmpty(t, items)
	menuItem = items[0]

	// f. Create order item, add customer as the owner
	orderItemReq := &proto.CreateOrderItemRequest{}
	orderItemReq.SetOrderId(order.GetId())
	orderItemReq.SetMenuItemId(menuItem.GetId())
	orderItemReq.SetQuantity(1)
	orderItemReq.SetCustomerOwnerIds([]string{cust.GetId()})
	orderItemID, err := orderClient.CreateOrderItem(ctx, orderItemReq)
	require.NoError(t, err)
	require.NotEmpty(t, orderItemID.GetId())

	// g. Send order
	sendOrderReq := &proto.SendOrderRequest{}
	sendOrderReq.SetOrderId(order.GetId())
	_, err = orderClient.SendOrder(ctx, sendOrderReq)
	require.NoError(t, err)

	openTab, err = tabClient.GetOpenTab(ctx, getTabReq)
	require.NoError(t, err)
	require.NotEmpty(t, openTab.GetId())
	order2 := openTab.GetOrders()[1]

	// h. Create guest
	createGuestReq := &proto.CreateGuestRequest{}
	createGuestReq.SetTabId(tabResp.GetId())
	guestIDResp, err := tabClient.CreateGuest(ctx, createGuestReq)
	require.NoError(t, err)
	require.NotEmpty(t, guestIDResp.GetId())

	// h. Update guest name
	updateGuestReq := &proto.UpdateGuestNameRequest{}
	updateGuestReq.SetGuestId(guestIDResp.GetId())
	_, err = tabClient.UpdateGuestName(ctx, updateGuestReq)
	require.NoError(t, err)

	// j. Create order item, add guest as the owner
	orderItemReq2 := &proto.CreateOrderItemRequest{}
	orderItemReq2.SetOrderId(order2.GetId())
	orderItemReq2.SetMenuItemId(menuItem.GetId())
	orderItemReq2.SetQuantity(1)
	orderItemReq2.SetGuestOwnerIds([]string{guestIDResp.GetId()})
	orderItemID2, err := orderClient.CreateOrderItem(ctx, orderItemReq2)
	require.NoError(t, err)
	require.NotEmpty(t, orderItemID2.GetId())

	// k. Add order item customer owners
	addOrderItemCustomerOwnersReq := &proto.AddOrderItemCustomerOwnerRequest{}
	addOrderItemCustomerOwnersReq.SetOrderItemId(orderItemID2.GetId())
	addOrderItemCustomerOwnersReq.SetCustomerId(cust.GetId())
	_, err = orderClient.AddOrderItemCustomerOwner(ctx, addOrderItemCustomerOwnersReq)
	require.NoError(t, err)

	// l. Remove order item customer owners
	removeOrderItemCustomerOwnersReq := &proto.RemoveOrderItemCustomerOwnerRequest{}
	removeOrderItemCustomerOwnersReq.SetOrderItemId(orderItemID2.GetId())
	removeOrderItemCustomerOwnersReq.SetCustomerId(cust.GetId())
	_, err = orderClient.RemoveOrderItemCustomerOwner(ctx, removeOrderItemCustomerOwnersReq)
	require.NoError(t, err)

	// m. Remove order item guest owners
	removeOrderItemGuestOwnersReq := &proto.RemoveOrderItemGuestOwnerRequest{}
	removeOrderItemGuestOwnersReq.SetOrderItemId(orderItemID2.GetId())
	removeOrderItemGuestOwnersReq.SetGuestId(guestIDResp.GetId())
	_, err = orderClient.RemoveOrderItemGuestOwner(ctx, removeOrderItemGuestOwnersReq)
	require.NoError(t, err)

	// n. Create order item, add guest as the owner
	orderItemReq3 := &proto.CreateOrderItemRequest{}
	orderItemReq3.SetOrderId(order2.GetId())
	orderItemReq3.SetMenuItemId(menuItem.GetId())
	orderItemReq3.SetQuantity(1)
	orderItemReq3.SetCustomerOwnerIds([]string{cust.GetId()})
	orderItemID3, err := orderClient.CreateOrderItem(ctx, orderItemReq3)
	require.NoError(t, err)
	require.NotEmpty(t, orderItemID3.GetId())

	// o. Update order item quantity
	updateOrderItemQtyReq := &proto.UpdateOrderItemQuantityRequest{}
	updateOrderItemQtyReq.SetOrderItemId(orderItemID3.GetId())
	updateOrderItemQtyReq.SetQuantity(2)
	_, err = orderClient.UpdateOrderItemQuantity(ctx, updateOrderItemQtyReq)
	require.NoError(t, err)

	// p. Add order item guest owners
	addOrderItemGuestOwnersReq := &proto.AddOrderItemGuestOwnerRequest{}
	addOrderItemGuestOwnersReq.SetOrderItemId(orderItemID3.GetId())
	addOrderItemGuestOwnersReq.SetGuestId(guestIDResp.GetId())
	_, err = orderClient.AddOrderItemGuestOwner(ctx, addOrderItemGuestOwnersReq)
	require.NoError(t, err)

	// q. Delete order item
	deleteOrderItemReq := &proto.DeleteOrderItemRequest{}
	deleteOrderItemReq.SetId(orderItemID2.GetId())
	_, err = orderClient.DeleteOrderItem(ctx, deleteOrderItemReq)
	require.NoError(t, err)

	// r. Send order
	sendOrderReq2 := &proto.SendOrderRequest{}
	sendOrderReq2.SetOrderId(order2.GetId())
	_, err = orderClient.SendOrder(ctx, sendOrderReq2)
	require.NoError(t, err)

	// s. Close tab
	closeTabReq := &proto.CloseTabRequest{}
	closeTabReq.SetTabId(tabResp.GetId())
	_, err = tabClient.CloseTab(ctx, closeTabReq)
	require.NoError(t, err)

	// t. Get closed tab
	openTab, err = tabClient.GetOpenTab(ctx, getTabReq)
	require.Error(t, err)
	require.Empty(t, openTab.GetId())

	// u. Get visited tab
	getVisitedTabsReq := &proto.GetVisitedTabsRequest{}
	getVisitedTabsReq.SetCustomerId(cust.GetId())
	visitedTabs, err := tabClient.GetVisitedTabs(ctx, getVisitedTabsReq, customerCred)
	require.NoError(t, err)
	require.NotEmpty(t, visitedTabs.GetTabs())
}

func GenerateSelfSignedTLSCert() (certPEM, keyPEM []byte, err error) {
	// Generate a private key (ECDSA P256)
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	// Create certificate template
	notBefore := time.Now()
	notAfter := notBefore.Add(time.Minute)
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	// Encode certificate to PEM
	certPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Encode private key to PEM
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}

	keyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privBytes,
	})

	return certPEM, keyPEM, nil
}
