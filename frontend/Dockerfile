# ==========================================
# Stage 1: Build Stage
# ==========================================
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package manifests first to leverage Docker layer caching
COPY package.json package-lock.json ./

# Install dependencies (ci ensures clean, deterministic builds)
RUN npm ci

# Copy the source code
COPY . .

# Build the production bundle (outputs to /app/dist or /app/build)
RUN npm run build

# ==========================================
# Stage 2: Production Stage (Nginx)
# ==========================================
FROM nginx:alpine

# Remove default Nginx website configuration
RUN rm -rf /etc/nginx/conf.d/default.conf

# Copy custom Nginx configuration for React SPA routing
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy compiled static assets from the build stage (adjust 'dist' to 'build' if using Create React App)
COPY --from=builder /app/dist /usr/share/nginx/html

# Expose HTTP port
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]