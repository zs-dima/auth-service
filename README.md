# Simple Go gRPC service

Simple Go gRPC service to discover and demonstrate best practices.
The service is based on the standard Go project layout: https://github.com/golang-standards/project-layout

## Features
- [x] JWT authentication
- [x] User management
- [ ] Event Bus to notify clients, probably Centrifugo
- [ ] Audit logging
- [ ] Test coverage
- [ ] Examples

Welcome to contribute, suggest and discuss improvements.

The service provides basic authentication and user management.

Real world applications will have more complex implementations:
- Roles: Roles management usually includes a hierarchy of roles and permission management.
- Audit: Audit logging to safely track user actions and changes.
- Passwords: Email verification, password reset, password strength requirements, etc.
- ...

It is a client application for testing the service:
https://github.com/zs-dima/auth-app


Audit logging can be based on a separate Time table or simple created_at/updated_at fields.
Let's leave the implementation out of scope for now, although you are welcome to discuss it as well.
