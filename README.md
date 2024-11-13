# Slugger - URL Shortener API

Slugger is a URL shortener backend application built with Go and Redis, containerized with Docker, and deployed on Heroku. It enables users to generate short URLs (or "slugs") and access the original URLs through a redirect.

## Features

- **Generate Short URLs**: Accepts a URL input and generates a unique short code or "slug" that can be shared easily.
- **Redirect to Original URL**: Accesses the original URL using the short code via a redirection endpoint.

## API Endpoints

### 1. **POST /shorten**

Generates a short code for a given URL by taking input via a form field named `url`.

- **Request**:
  - Method: `POST`
  - URL: `/shorten`
  - Content-Type: `application/x-www-form-urlencoded`
  - Form Data:
    - `url`: The URL to shorten, e.g., `https://example.com`

- **Response**:
  - Success:
    ```json
    {
      "shortUrl": "https://yourapp.com/r/{shortCode}"
    }
    ```
  - Error:
    ```json
    {
      "error": "Error message"
    }
    ```

### 2. **GET /r/{shortCode}**

Redirects to the original URL associated with the given short code.

- **Request**:
  - Method: `GET`
  - URL: `/r/{shortCode}`

- **Response**:
  - Success: Redirects the user to the original URL.
  - Error: Returns a 404 status if the short code is not found.

## Technologies Used

- **Backend**: Go
- **Database**: Redis
- **Containerization**: Docker
- **Deployment**: Heroku
