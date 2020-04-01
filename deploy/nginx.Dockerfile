FROM nginx
COPY nginx.conf /etc/nginx
RUN mkdir -p /etc/nginx/html/.well-known/acme-challenge/ && touch YX0BIAnEyrV0gqgG-ZANN2dFiAA9TW53ver-w1pbzvYp5Mqq6iFlScfk6FHchaF