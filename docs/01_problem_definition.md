
# Problem Definition

The goal of this project is to design and implement a scalable URL shortening service similar to Bitly.

A URL shortener converts long URLs into shorter ones that are easier to share and manage.

Example:

Long URL:
https://example.com/some/very/long/path

Short URL:
https://short.ly/abc123

When users access the short URL, the system redirects them to the original long URL.

This project will implement the system using a polyglot microservices architecture:

- Go for API Gateway
- Python for URL service
- Java for analytics service

The project will focus on understanding system design tradeoffs including:

- ID generation
- caching
- scalability
- distributed services