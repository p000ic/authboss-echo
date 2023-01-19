# Authboss Echo

A carbon copy from

[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4)](https://pkg.go.dev/mod/github.com/p000ic/authboss-echo)
![ActionsCI](https://github.com/p000ic/authboss-echo/workflows/test/badge.svg)
[![Mail](https://img.shields.io/badge/mail%20list-authboss-lightgrey.svg)](https://groups.google.com/a/volatile.tech/forum/#!forum/authboss)
![Go Coverage](https://img.shields.io/badge/coverage-84.9%25-brightgreen.svg?longCache=true&style=flat)
[![Go Report Card](https://goreportcard.com/badge/github.com/p000ic/authboss-echo)](https://goreportcard.com/report/github.com/p000ic/authboss-echo)

Authboss is a modular authentication system for the web.

It has several modules that represent authentication and authorization features that are common
to websites in general so that you can enable as many as you need, and leave the others out.
It makes it easy to plug in authentication to an application and get a lot of functionality
for (hopefully) a smaller amount of integration effort.

# New to v2?

v1 -> v2 was a very big change. If you're looking to upgrade there is a general guide in
[tov2.md](tov2.md) in this project.

# New to v3?

v2 -> v3 was not a big change, it simply changed the project to use Go modules.
Authboss no longer supports GOPATH as of version 3

# Why use Authboss?

Every time you'd like to start a new web project, you really want to get to the heart of what you're
trying to accomplish very quickly, and it would be a sure bet to say one of the systems you're excited
about implementing and innovating on is not authentication. In fact, it's very much the opposite: it's
one of those things that you have to do and one of those things you loathe to do. Authboss is supposed
to remove a lot of the tedium that comes with this, as well as a lot of the chances to make mistakes.
This allows you to care about what you're intending to do, rather than care about ancillary support
systems required to make what you're intending to do happen.

Here are a few bullet point reasons you might like to try it out:

* Saves you time (Authboss integration time should be less than re-implementation time)
* Saves you mistakes (at least using Authboss, people can bug fix as a collective and all benefit)
* Should integrate with or without any web framework

# [Click Here To Get Started](https://volatiletech.github.io/authboss/#/migration)

# Readme Table of Contents
<!-- TOC -->

- [Authboss](#authboss)
- [New to v2?](#new-to-v2)
- [New to v3?](#new-to-v3)
- [Why use Authboss?](#why-use-authboss)
- [Readme Table of Contents](#readme-table-of-contents)
- [Getting Started](#getting-started)
    - [App Requirements](#app-requirements)
        - [CSRF Protection](#csrf-protection)
        - [Request Throttling](#request-throttling)
    - [Integration Requirements](#integration-requirements)
        - [Middleware](#middleware)
        - [Configuration](#configuration)
        - [Storage and Core implementations](#storage-and-core-implementations)
        - [ServerStorer implementation](#serverstorer-implementation)
        - [User implementation](#user-implementation)
        - [Values implementation](#values-implementation)
    - [Config](#config)
        - [Paths](#paths)
        - [Modules](#modules)
        - [Mail](#mail)
        - [Storage](#storage)
        - [Core](#core)
- [Available Modules](#available-modules)
- [Middlewares](#middlewares)
- [Use Cases](#use-cases)
    - [Get Current User](#get-current-user)
    - [Reset Password](#reset-password)
    - [User Auth via Password](#user-auth-via-password)
    - [User Auth via OAuth1](#user-auth-via-oauth1)
    - [User Auth via OAuth2](#user-auth-via-oauth2)
    - [User Registration](#user-registration)
    - [Confirming Registrations](#confirming-registrations)
    - [Password Recovery](#password-recovery)
    - [Remember Me](#remember-me)
    - [Locking Users](#locking-users)
    - [Expiring User Sessions](#expiring-user-sessions)
    - [One Time Passwords](#one-time-passwords)
    - [Two Factor Authentication](#two-factor-authentication)
        - [Two-Factor Recovery](#two-factor-recovery)
        - [Two-Factor Setup E-mail Authorization](#two-factor-setup-e-mail-authorization)
        - [Time-Based One Time Passwords 2FA (totp)](#time-based-one-time-passwords-2fa-totp)
            - [Adding 2fa to a user](#adding-2fa-to-a-user)
            - [Removing 2fa from a user](#removing-2fa-from-a-user)
            - [Logging in with 2fa](#logging-in-with-2fa)
            - [Using Recovery Codes](#using-recovery-codes)
        - [Text Message 2FA (sms)](#text-message-2fa-sms)
            - [Adding 2fa to a user](#adding-2fa-to-a-user-1)
            - [Removing 2fa from a user](#removing-2fa-from-a-user-1)
            - [Logging in with 2fa](#logging-in-with-2fa-1)
            - [Using Recovery Codes](#using-recovery-codes-1)
    - [Rendering Views](#rendering-views)
        - [HTML Views](#html-views)
        - [JSON Views](#json-views)
        - [Data](#data)

<!-- /TOC -->

# Getting Started

To get started with Authboss in the simplest way, is to simply create a Config, populate it
with the things that are required, and start implementing [use cases](#use-cases). The use
cases describe what's required to be able to use a particular piece of functionality,
or the best practice when implementing a piece of functionality. Please note the
[app requirements](#app-requirements) for your application as well
[integration requirements](#integration-requirements) that follow.

Of course the standard practice of fetching the library is just the beginning:

```bash
# Get the latest, you must be using Go modules as of v3 of Authboss.
go get -u github.com/p000ic/authboss-echo
```

Here's a bit of starter code that was stolen from the sample.

```go
ab := authboss.New()

ab.Config.Storage.Server = myDatabaseImplementation
ab.Config.Storage.SessionState = mySessionImplementation
ab.Config.Storage.CookieState = myCookieImplementation

ab.Config.Paths.Mount = "/authboss"
ab.Config.Paths.RootURL = "https://www.example.com/"

// This is using the renderer from: github.com/p000ic/authboss-echo
ab.Config.Core.ViewRenderer = abrenderer.NewHTML("/auth", "ab_views")
// Probably want a MailRenderer here too.


// This instantiates and uses every default implementation
// in the Config.Core area that exist in the defaults package.
// Just a convenient helper if you don't want to do anything fancy.
 defaults.SetCore(&ab.Config, false, false)

if err := ab.Init(); err != nil {
    panic(err)
}

// Mount the router to a path (this should be the same as the Mount path above)
// mux in this example is a chi router, but it could be anything that can route to
// the Core.Router.
mux.Mount("/authboss", http.StripPrefix("/authboss", ab.Config.Core.Router))
```

For a more in-depth look you **definitely should** look at the authboss sample to see what a full
implementation looks like. This will probably help you more than any of this documentation.

[https://github.com/p000ic/authboss-echo-sample](https://github.com/p000ic/authboss-echo-sample)

## App Requirements

Authboss does a lot of things, but it doesn't do some of the important things that are required by
a typical authentication system, because it can't guarantee that you're doing many of those things
in a different way already, so it punts the responsibility.

### CSRF Protection

What this means is you should apply a middleware that can protect the application from csrf
attacks or you may be vulnerable. Authboss previously handled this but it took on a dependency
that was unnecessary and it complicated the code. Because Authboss does not render views nor
consumes data directly from the user, it no longer does this.

### Request Throttling

Currently Authboss is vulnerable to brute force attacks because there are no protections on
it's endpoints. This again is left up to the creator of the website to protect the whole website
at once (as well as Authboss) from these sorts of attacks.

## Integration Requirements

In terms of integrating Authboss into your app, the following things must be considered.

### Middleware

There are middlewares that are required to be installed in your middleware stack if it's
all to function properly, please see [Middlewares](#middlewares) for more information.

### Configuration

There are some required configuration variables that have no sane defaults and are particular
to your app:

* Config.Paths.Mount
* Config.Paths.RootURL

### Storage and Core implementations

Everything under Config.Storage and Config.Core are required and you must provide them,
however you can optionally use default implementations from the
[defaults package](https://github.com/p000ic/authboss-echo/tree/master/defaults).
This also provides an easy way to share implementations of certain stack pieces (like HTML Form Parsing).
As you saw in the example above these can be easily initialized with the `SetCore` method in that
package.

The following is a list of storage interfaces, they must be provided by the implementer. Server is a
very involved implementation, please see the additional documentation below for more details.

* Config.Storage.Server
* Config.Storage.SessionState
* Config.Storage.CookieState (only for "remember me" functionality)

The following is a list of the core pieces, these typically are abstracting the HTTP stack.
Out of all of these you'll probably be mostly okay with the default implementations in the
defaults package but there are two big exceptions to this rule and that's the ViewRenderer
and the MailRenderer. For more information please see the use case [Rendering Views](#rendering-views)

* Config.Core.Router
* Config.Core.ErrorHandler
* Config.Core.Responder
* Config.Core.Redirector
* Config.Core.BodyReader
* Config.Core.ViewRenderer
* Config.Core.MailRenderer
* Config.Core.Mailer
* Config.Core.Logger

### ServerStorer implementation

The [ServerStorer](https://pkg.go.dev/github.com/p000ic/authboss-echo/#ServerStorer) is
meant to be upgraded to add capabilities depending on what modules you'd like to use.
It starts out by only knowing how to save and load users, but the `remember` module as an example
needs to be able to find users by remember me tokens, so it upgrades to a
[RememberingServerStorer](https://pkg.go.dev/github.com/p000ic/authboss-echo/#RememberingServerStorer)
which adds these abilities.

Your `ServerStorer` implementation does not need to implement all these additional interfaces
unless you're using a module that requires it. See the [Use Cases](#use-cases) documentation to know what the requirements are.

### User implementation

Users in Authboss are represented by the
[User interface](https://pkg.go.dev/github.com/p000ic/authboss-echo/#User). The user
interface is a flexible notion, because it can be upgraded to suit the needs of the various modules.

Initially the User must only be able to Get/Set a `PID` or primary identifier. This allows the authboss
modules to know how to refer to him in the database. The `ServerStorer` also makes use of this
to save/load users.

As mentioned, it can be upgraded, for example suppose now we want to use the `confirm` module,
in that case the e-mail address now becomes a requirement. So the `confirm` module will attempt
to upgrade the user (and panic if it fails) to a
[ConfirmableUser](https://pkg.go.dev/github.com/p000ic/authboss-echo/#ConfirmableUser)
which supports retrieving and setting of confirm tokens, e-mail addresses, and a confirmed state.

Your `User` implementation does not need to implement all these additional user interfaces unless you're
using a module that requires it. See the [Use Cases](#use-cases) documentation to know what the
requirements are.

### Values implementation

The [BodyReader](https://pkg.go.dev/github.com/p000ic/authboss-echo/#BodyReader)
interface in the Config returns
[Validator](https://pkg.go.dev/github.com/p000ic/authboss-echo/#Validator) implementations
which can be validated. But much like the storer and user it can be upgraded to add different
capabilities.

A typical `BodyReader` (like the one in the defaults package) implementation checks the page being
requested and switches on that to parse the body in whatever way
(msgpack, json, url-encoded, doesn't matter), and produce a struct that has the ability to
`Validate()` it's data as well as functions to retrieve the data necessary for the particular
valuer required by the module.

An example of an upgraded `Valuer` is the
[UserValuer](https://pkg.go.dev/github.com/p000ic/authboss-echo/#UserValuer)
which stores and validates the PID and Password that a user has provided for the modules to use.

Your body reader implementation does not need to implement all valuer types unless you're
using a module that requires it. See the [Use Cases](#use-cases) documentation to know what the
requirements are.

## Config

The config struct is an important part of Authboss. It's the key to making Authboss do what you
want with the implementations you want. Please look at it's code definition as you read the
documentation below, it will make much more sense.

[Config Struct Documentation](https://pkg.go.dev/github.com/p000ic/authboss-echo/#Config)

### Paths

Paths are the paths that should be redirected to or used in whatever circumstance they describe.
Two special paths that are required are `Mount` and `RootURL` without which certain authboss
modules will not function correctly. Most paths get defaulted to `/` such as after login success
or when a user is locked out of their account.

### Modules

Modules are module specific configuration options. They mostly control the behavior of modules.
For example `RegisterPreserveFields` decides a whitelist of fields to allow back into the data
to be re-rendered so the user doesn't have to type them in again.

### Mail

Mail sending related options.

### Storage

These are the implementations of how storage on the server and the client are done in your
app. There are no default implementations for these at this time. See the [Godoc](https://pkg.go.dev/mod/github.com/p000ic/authboss-echo) for more information
about what these are.

### Core

These are the implementations of the HTTP stack for your app. How do responses render? How are
they redirected? How are errors handled?

For most of these there are default implementations from the
[defaults package](https://github.com/p000ic/authboss-echo/tree/master/defaults) available, but not for all.
See the package documentation for more information about what's available.

# Available Modules

Each module can be turned on simply by importing it and the side-effects take care of the rest.
Not all the capabilities of authboss are represented by a module, see [Use Cases](#use-cases)
to view the supported use cases as well as how to use them in your app.

**Note**: The two factor packages do not enable via side-effect import, see their documentation
for more information.

| Name      | Import Path                                           | Description                                                    |
|-----------|-------------------------------------------------------|----------------------------------------------------------------|
| Auth      | github.com/p000ic/authboss-echo/auth                  | Database password authentication for users.                    |
| Confirm   | github.com/p000ic/authboss-echo/confirm               | Prevents login before e-mail verification.                     |
| Expire    | github.com/p000ic/authboss-echo/expire                | Expires a user's login                                         |
| Lock      | github.com/p000ic/authboss-echo/lock                  | Locks user accounts after authentication failures.             |
| Logout    | github.com/p000ic/authboss-echo/logout                | Destroys user sessions for auth/oauth2.                        |
| OAuth1    | github.com/epiphenomena/authboss-oauth1               | Provides oauth1 authentication for users.                      |
| OAuth2    | github.com/p000ic/authboss-echo/oauth2                | Provides oauth2 authentication for users.                      |
| Recover   | github.com/p000ic/authboss-echo/recover               | Allows for password resets via e-mail.                         |
| Register  | github.com/p000ic/authboss-echo/register              | User-initiated account creation.                               |
| Remember  | github.com/p000ic/authboss-echo/remember              | Persisting login sessions past session cookie expiry.          |
| OTP       | github.com/p000ic/authboss-echo/otp                   | One time passwords for use instead of passwords.               |
| Twofactor | github.com/p000ic/authboss-echo/otp/twofactor         | Regenerate recovery codes for 2fa.                             |
| Totp2fa   | github.com/p000ic/authboss-echo/otp/twofactor/totp2fa | Use Google authenticator-like things for a second auth factor. |
| Sms2fa    | github.com/p000ic/authboss-echo/otp/twofactor/sms2fa  | Use a phone for a second auth factor.                          |

# Middlewares

The only middleware that's truly required is the `LoadClientStateMiddleware`, and that's because it
enables session and cookie handling for Authboss. Without that, it's not a very useful piece of
software.

The remaining middlewares are either the implementation of an entire module (like expire),
or a key part of a module. For example you probably wouldn't want to use the lock module
without the middleware that would stop a locked user from using an authenticated resource,
because then locking wouldn't be useful unless of course you had your own way of dealing
with locking, which is why it's only recommended, and not required. Typically you will
use the middlewares if you use the module.

| Name                                                                                                                | Requirement               | Description                                           |
|---------------------------------------------------------------------------------------------------------------------|---------------------------|-------------------------------------------------------|
| [Middleware](https://pkg.go.dev/github.com/p000ic/authboss-echo/#Middleware)                                        | Recommended               | Prevents unauthenticated users from accessing routes. |
| [LoadClientStateMiddleware](https://pkg.go.dev/github.com/p000ic/authboss-echo/#Authboss.LoadClientStateMiddleware) | **Required**              | Enables cookie and session handling                   |
| [ModuleListMiddleware](https://pkg.go.dev/github.com/p000ic/authboss-echo/#Authboss.ModuleListMiddleware)           | Optional                  | Inserts a loaded module list into the view data       |
| [confirm.Middleware](https://pkg.go.dev/github.com/p000ic/authboss-echo/confirm/#Middleware)                        | Recommended with confirm  | Ensures users are confirmed or rejects request        |
| [expire.Middleware](https://pkg.go.dev/github.com/p000ic/authboss-echo/expire/#Middleware)                          | **Required** with expire  | Expires user sessions after an inactive period        |
| [lock.Middleware](https://pkg.go.dev/github.com/p000ic/authboss-echo/lock/#Middleware)                              | Recommended with lock     | Rejects requests from locked users                    |
| [remember.Middleware](https://pkg.go.dev/github.com/p000ic/authboss-echo/remember/#Middleware)                      | Recommended with remember | Logs a user in from a remember cookie                 |



## Rendering Views

The authboss rendering system is simple. It's defined by one interface: [Renderer](https://pkg.go.dev/github.com/p000ic/authboss-echo/#Renderer)

The renderer knows how to load templates, and how to render them with some data and that's it.
So let's examine the most common view types that you might want to use.

### HTML Views

When your app is a traditional web application and is generating its HTML
serverside using templates this becomes a small wrapper on top of your rendering
setup. For example if you're using `html/template` then you could just use
`template.New()` inside the `Load()` method and store that somewhere and call
`template.Execute()` in the `Render()` method.

There is also a very basic renderer: [Authboss
Renderer](https://github.com/p000ic/authboss-echo-renderer) which has some very
ugly built in views and the ability to override them with your own if you don't
want to integrate your own rendering system into that interface.

### JSON Views

If you're building an API that's mostly backed by a javascript front-end, then you'll probably
want to use a renderer that converts the data to JSON. There is a simple json renderer available in
the [defaults package](https://github.com/p000ic/authboss-echo/tree/master/defaults) if you wish to
use that.

### Data

The most important part about this interface is the data that you have to render.
There are several keys that are used throughout authboss that you'll want to render in your views.

They're in the file [html_data.go](https://github.com/p000ic/authboss-echo/blob/master/html_data.go)
and are constants prefixed with `Data`. See the documentation in that file for more information on
which keys exist and what they contain.

The default [responder](https://pkg.go.dev/github.com/p000ic/authboss-echo/defaults/#Responder)
also happens to collect data from the Request context, and hence this is a great place to inject
data you'd like to render (for example data for your html layout, or csrf tokens).
