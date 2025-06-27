# Entities

## Customer

### Description

The person dining in.

### Attributes

- ID
- Name
- Contact Info

## Menu Item

### Description

Individual food or drink items available for ordering.

### Attributes

- ID
- Name
- Description
- Photo
- Price
- Portion Size
- Availability
- Modifiers Configuration
- Menu Tags

## Menu Tag

### Description

A label attached to menu item for the purpose of categorization.

### Attributes

- ID
- Value
- Description
- Menu Tag Dimension ID
- Prerequisites

## Menu Tag Dimension

### Description

A way to categorize a menu tag.

### Attributes

- ID
- Value
- Description

## Order

### Description

A grouping of order items that is sent together to the kitchen.

### Attributes

- ID
- Tab ID
- Sent At

## Order Item

### Description

A line item in an order, linking a menu item to a specific order with owners, quantity, and modifiers.

### Attributes

- ID
- Order ID
- Menu Item ID
- Owner ID(s)
- Quantity
- Modifiers

## Tab

### Description

A session that will track everything customers order during their visit.

### Attributes

- ID
- Total Price
- Created At
- Closed At

# Features

## Arrival

### QR Code

When a customer comes to the restaurant,
the backend server will create a new tab.
The customer will scan a QR code containing the tab ID that will open a link to the restaurant website.
The QR code can be scanned multiple times from different devices.

### Unregistered Customer

Customer doesn't have to have an account in order to use the app.
Customer may optionally create an account to track their visits.
When a customer use the app without an account,
they will be identified by a temporary id that's unique within the tab.
Each device that doesn't use an account will be treated as different customer.

#### Customer Name

##### Default Name

The first time the restaurant website is opened for this tab on each device,
the device will be assigned with a random name consisting of an positive adjective and an animal.

The name should be unique for each unregistered customer within the tab.
1. If the number of unregistered customer is less or equal than the number of available animal,
then the assigned animal should be unique and the assigned adjective should be unique.
2. If the number of unregistered customer is less or equal than the number of combination of available animal and adjective,
then the assigned animal + adjective should be unique.
3. Else, append number to the name.

```
<examples>
  <animals>
    <animal>Elephant</animal>
    <animal>Tiger</animal>
    <animal>Dolphin</animal>
  </animals>

  <adjectives>
    <adjective>Cute</adjective>
    <adjective>Smart</adjective>
    <adjective>Strong</adjective>
  </adjectives>

  <example case=1>
    <customer id=1 registered="unregistered" name="Smart Elephant" />
    <customer id=2 registered="unregistered" name="Strong Tiger" />
    <customer id=3 registered="unregistered" name="Cute Dolphin" />
  </example>

  <example case=2>
    <customer id=1 registered="unregistered" name="Smart Elephant" />
    <customer id=2 registered="unregistered" name="Strong Tiger" />
    <customer id=3 registered="unregistered" name="Cute Dolphin" />
    <customer id=4 registered="unregistered" name="Cute Tiger" />
    <customer id=5 registered="unregistered" name="Strong Elephant" />
    <customer id=6 registered="unregistered" name="Smart Dolphin" />
    <customer id=7 registered="unregistered" name="Smart Tiger" />
    <customer id=8 registered="unregistered" name="Cute Elephant" />
    <customer id=9 registered="unregistered" name="Strong Dolphin" />
  </example>

  <example case=3>
    <customer id=1 registered="unregistered" name="Smart Elephant" />
    <customer id=2 registered="unregistered" name="Strong Tiger" />
    <customer id=3 registered="unregistered" name="Cute Dolphin" />
    <customer id=4 registered="unregistered" name="Cute Tiger" />
    <customer id=5 registered="unregistered" name="Strong Elephant" />
    <customer id=6 registered="unregistered" name="Smart Dolphin" />
    <customer id=7 registered="unregistered" name="Smart Tiger" />
    <customer id=8 registered="unregistered" name="Cute Elephant" />
    <customer id=9 registered="unregistered" name="Strong Dolphin" />
    <customer id=10 registered="unregistered" name="Smart Tiger 2" />
  </example>
</examples>
```

##### Custom Name

The customer may update the default name to a custom name.
The customer name will be stored on the device so it can be reused again in the future.

## Browsing Menu

The customer can browse the menu and use filtering based on the menu tag.

## Ordering

Only customer who know the Tab ID can modify the order.

### Adding Items

The customer can add items with modifiers to the current order.

### Updating Items

The customer can update items in the current order.

#### Sharing Items

A customer can share the same item with another customer in the same tab by adding themself to the item's owners.

##### Unsharing Items

A customer can remove themself from the item's owner.
If the customer is the only item's owner, the request will be rejected.

### Removing Items

The customer can remove items from the current order.

### Sending Order

The customer can send order multiple times to the kitchen.
Once an order is sent, the customer can only update the owner of the order item, and a new empty order is created again.

## Payment

The customer can trigger the payment process after ordering.
The backend service will calculate the total amount and generate a QRIS code.
The tab will be closed when backend service confirmed the payment.
Once a tab is closed, the customer cannot edit any items in the tab.
