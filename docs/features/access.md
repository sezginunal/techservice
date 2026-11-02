# Access

Status: approved.

## Login

- The unauthenticated visitor sees only the login screen.
- The screen collects a username and a password and can switch the interface between English and Turkish.
- The language choice is stored in the browser and applies before login.
- A failed login says that the username or password is wrong. It does not reveal which one.
- A successful login opens a home screen that shows the signed-in name and role.
- An admin also gets a way to open user management. A technician or an owner does not.

## Session

- A successful login sets an HTTP-only session cookie.
- The server stores the session in SQLite with an absolute expiry of 7 days from login. Activity does not extend it.
- An expired session is deleted and the user must log in again.
- Each request loads the session and the user.
- Logout deletes that session and clears the cookie.
- Deactivating a user or changing that user's password deletes that user's sessions.
- A missing, expired, or inactive user gets the login screen. The API responds with unauthorized.

## Users

- Roles are `technician`, `owner`, and `admin`.
- Only an admin can list users, create a user, change a role, set a password, or deactivate a user.
- A create asks for the username, role, and initial password.
- A password is at least 8 characters. A shorter password is rejected.
- The last active admin cannot be deactivated or changed to another role.
- Usernames are unique.
- Passwords are stored with bcrypt (`golang.org/x/crypto/bcrypt`). The plain password is never stored or logged.

## First admin

- `ADMIN_USERNAME` and `ADMIN_PASSWORD` come from the environment.
- When the database has no admin, the backend creates one from those values and stores the password hash.
- When an admin already exists, the backend leaves the stored user unchanged.
- When either value is missing, or the password is shorter than 8 characters, and no admin exists, the backend refuses to start. The log names the missing or invalid variable and does not include the password.
