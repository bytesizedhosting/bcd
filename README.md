## Bytesized Connect Daemon (BCD)

This is the daemon part of the Bytesized Connect project running at connect.bytesized-hosting.com.

Bytesized Connect transforms your (virtual) server into a application platform that you can manage from the Bytesized website.

Everything in this repository is pretty much proof-of-concept if you use this be ready to update your software _a lot_.

### What is Bytesized Connect?

Bytesized Connect is a control panel that lets you install any supported applications without any command line configuration through the Bytesized website. It also displays and graphs information about your resources. And can generate unique rememberable domain names for easy access.

__Connect Overview__
![Connect Overview](https://dl.dropboxusercontent.com/u/374/engine/engine_overview.png)

### Installation

#### Automatic installation
If you are on Ubuntu 14.04 or 16.04 you can use the automatic installer. To install simply create a new Connect Account on the [Bytesized website](https://bytesized-hosting.com/engine/accounts/new) and follow the steps.

#### Manual installation
These are the basic steps required to run BCD without the automatic installer.

* Grab the latest release from the releases page.
* Install Docker.
* Create a user that will run the deamon.
* Run `bcd init <apikey> <apisecret>`
* Run `bysh-engine start`

If you want to support for proxies using bcd-proxy make sure you install
that as well.

### Architecture

Bytesized Connect exists of two pieces of software, the Deamon (BCD), who's code you are looking at now, and the Connect Manager, which can be found on https://bytesized-hosting.com

The overall architecture of the Bytesized Connect was designed with isolation as the most important feature. This translates in the following features:

* All apps are run inside of Docker containers;
* BCD can run without any sudo privileges;
* All filesystem operations are isolated to one user;

This means that it should always be safe to run Bytesized Connect next to any other processes without causing any interferance and you can use whatever modules you prefer and add them onto existing software.

#### Plugins

Plugins form the heart of Bytesized Connect. Every app we want to support should have a custom plugin with it's own installation logic.

There are two types of plugins

* Discoverable (App) plugins
* Feature plugins

##### Feature plugins

Feature plugins add features to the Bytesized Connect platform. These plugins are not automatically discoverable and support has to be added manually by Bytesized.

The Stats plugin is an example of a feature plugin. It adds support for various system information and is hard coded into the Connect Manager.

##### Discoverable (App) plugins

Discoverable, or App, plugins are a special type of Plugin which can be automatically used from Bytesized Connect without any extra coding required on the manager itself.

It achieves this by conforming to a standard.

* Every App plugin runs inside a Docker container;
* Every App has a manifest describing how Bytesized Connect should install and interact with the application;

Here is an example of a manifest

```go
exposed_methods:
- Install
- Restart
- Stop
- Start
method_options:
  Install:
  - default_value: ""
    hint: ""
    name: password
    type: string
  - default_value: bytesized
    hint: ""
    name: username
    type: string
  - default_value: /home/bytesized/config/deluge
    hint: ""
    name: config_folder
    type: string
  - default_value: /home/bytesized/data
    hint: ""
    name: data_folder
    type: string
  Restart:
  - default_value: ""
    hint: ""
    name: container_id
    type: string
  Start:
  - default_value: ""
    hint: ""
    name: container_id
    type: string
  Stop:
  - default_value: ""
    hint: ""
    name: container_id
    type: string
name: Deluge
rpc_name: DelugeRPC
show_options:
- username
- password
- web_port
- daemon_port
- config_folder
- data_folder
version: 1
web_url_format: http://##ip##:##web_port##/
```

and here is how this manifest translates to the manager.

_The install options_
![Install options](https://dl.dropboxusercontent.com/u/374/engine/engine_install.png)

_And once installed_
![App details](https://dl.dropboxusercontent.com/u/374/engine/engine_installed.png)

This enables you to run any app you want to run without having to wait for Bytesized to add support for it.

### Developing an App Plugin

To create your own app you require the following things:

1. A Docker container of the app.

This container should try to adhere as much to the Bytesized Connect
standards as possible. This means it expects `/data`, `/config` and a
`/media` volume when it makes sense. The container should be as small as
possible, with a preference for Alpine. Files should be owned by the
right UID and GUID. See the bytesized/base images for more info.

2. A BCD Plugin

By installing `bcd/cmd/bcd-generate` you can create a simple skeleton
structure. For instance `bcd-generate plugin madsonic` to create a
plugin with the madsonic name. Connect images should always try to
create user authentication when a sane method to create a user exists
within the software. Some software let's you create users via a config
file, others come with an API. In some cases though it requires direct
database manipulation which might not be worth the effort.

Don't forget to load your plugin in main.go to make sure it's started.
