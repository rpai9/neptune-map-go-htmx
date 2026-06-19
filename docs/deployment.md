# Deployment Documentation

This is the deployment document for [Project Name], a Go + HTMX project.

## Prerequisites

Make sure you have Go installed on your system (version 1.15 or higher)
Ensure that your environment meets the following requirements:

- Go compiler
- HTMX library installed
- Web server software (e.g., Nginx, Apache)

## Steps to Deploy

- Clone the repository using git clone https://github.com/rpai/neptune-map-go-htmx
- Navigate to the project directory and run go mod tidy to ensure dependencies are up-to-date
- Run go build main.go to compile the Go code
- Create a new directory for your web server configuration files (e.g., /etc/nginx/sites-available/)
- Configure Nginx or Apache to serve your Go application by creating a new file in the directory above (e.g., nginx.conf) with the following contents:

```bash

server {
    listen 80;
    server_name [your_domain.com];

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

- Restart your web server software to apply the changes
- Open a web browser and navigate to http://[your_domain.com] to test your application.
