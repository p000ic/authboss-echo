# Available Modules

Each module can be turned on simply by importing it and the side effects take care of the rest.
Not all the capabilities of authboss are represented by a module, see [Use Cases](#use-cases)
to view the supported use cases as well as how to use them in your app.

**Note**: The two factor packages do not enable via side effect import, see their documentation
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
