---
applyTo: "**"
---
# Project general coding standards

## Philosophy
- Use functional programming imperative shell principles where possible
- Use monolithic architecture for simplicity
- Write secure code
- Write code that is easy to test
- Write performant code
- Write code that is easy to extend
- Avoid hardcoding values

## Dependency
- Use standard library where possible
- Use third-party packages when they are well-maintained and widely used

## Entry Point
- Have a CLI entry point
- Have a server entry point
- Have a TUI entry point

## Directory Structure

### Go Directories

#### `/cmd`

Main applications for this project.

The directory name for each application should match the name of the executable you want to have (e.g., `/cmd/myapp`).

Don't put a lot of code in the application directory.
If you think the code can be imported and used in other projects,
then it should live in the `/pkg` directory.
If the code is not reusable or if you don't want others to reuse it,
put that code in the `/internal` directory.
Be explicit about your intentions!

It's common to have a small `main` function that imports and invokes the code from the `/internal` and `/pkg` directories and nothing else.

#### `/internal`

Private application and library code.
This is the code you don't want others importing in their applications or libraries.
Note that you are not limited to the top level `internal` directory.
You can have more than one `internal` directory at any level of your project tree.

You can optionally add a bit of extra structure to your internal packages to separate your shared and non-shared internal code.
It's not required (especially for smaller projects),
but it's nice to have visual clues showing the intended package use.
Your actual application code can go in the `/internal/app` directory (e.g., `/internal/app/myapp`)
and the code shared by those apps in the `/internal/pkg` directory (e.g., `/internal/pkg/myprivlib`).

You use internal directories to make packages private.
If you put a package inside an internal directory,
then other packages can’t import it unless they share a common ancestor.

#### `/pkg`

Library code that's ok to use by external applications (e.g., `/pkg/mypubliclib`).
Other projects will import these libraries expecting them to work, so think twice before you put something here.
Note that the `internal` directory is a better way to ensure your private packages are not importable because it's enforced by Go.
The `/pkg` directory is still a good way to explicitly communicate that the code in that directory is safe for use by others.

### Service Application Directories

#### `/api`

OpenAPI/Swagger specs, JSON schema files, protocol definition files.

### Common Application Directories

#### `/configs`

Configuration file templates or default configs.

Put your `confd` or `consul-template` template files here.

#### `/init`

System init (systemd, upstart, sysv) and process manager/supervisor (runit, supervisord) configs.

#### `/scripts`

Scripts to perform various build, install, analysis, etc operations.

#### `/build`

Packaging and Continuous Integration.

Put your cloud (AMI), container (Docker), OS (deb, rpm, pkg) package configurations and scripts in the `/build/package` directory.

Put your CI (travis, circle, drone) configurations and scripts in the `/build/ci` directory.
Note that some of the CI tools (e.g., Travis CI) are very picky about the location of their config files.
Try putting the config files in the `/build/ci` directory linking them to the location where the CI tools expect them (when possible).

#### `/deployments`

IaaS, PaaS, system and container orchestration deployment configurations and templates (docker-compose, kubernetes/helm, terraform).
Note that in some repos (especially apps deployed with kubernetes) this directory is called `/deploy`.

#### `/test`

Additional external test apps and test data. Feel free to structure the `/test` directory anyway you want.
For bigger projects it makes sense to have a data subdirectory.
For example, you can have `/test/data` or `/test/testdata` if you need Go to ignore what's in that directory.
Note that Go will also ignore directories or files that begin with "." or "_", so you have more flexibility in terms of how you name your test data directory.

### Other Directories

#### `/docs`

Design and user documents (in addition to your godoc generated documentation).

#### `/tools`

Supporting tools for this project. Note that these tools can import code from the `/pkg` and `/internal` directories.

#### `/examples`

Examples for your applications and/or public libraries.

#### `/third_party`

External helper tools, forked code and other 3rd party utilities (e.g., Swagger UI).

#### `/githooks`

Git hooks.
