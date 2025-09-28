# Tourist Application (SOA Project)

## Overview
This project is a microservices-based **Tourist Application** designed for managing users, tours, blogs, and interactions in a tourism environment.  
It demonstrates the use of **Docker**, **NoSQL databases**, **microservices architecture**, and **RPC communication**.  

The system supports multiple user roles: **Administrator**, **Tour Guide**, and **Tourist**.  

The application is fully containerized and can be started using `docker-compose`, with monitoring and logging included.

**Note:** The Angular frontend is hosted in a separate repository.

---

## Frontend Repository
The Angular frontend for this application is hosted in a separate repository:  
[Tourist Application Frontend](https://github.com/akoskissak/soa-team-5-frontend)  

---

## Technologies Used
- **Backend:** Golang, Python
- **Frontend:** Angular
- **Databases:** NoSQL (document and graph databases such as MongoDB, Neo4j)  
- **Communication:** REST, RPC  
- **Containerization & Orchestration:** Docker, Docker Compose  
- **Monitoring & Logging:** Tracing, log aggregation, system and container metrics  

---

## Microservices & Architecture
The application is divided into several microservices:

| Service | Responsibilities |
|---------|-----------------|
| **Stakeholders Service** | User registration, authentication, and role management, User profile management, admin functionalities |
| **Blog Service** | Create and manage blogs, comments, likes |
| **Purchase Service** | Shopping cart, tour purchase, order tracking |
| **Followers Service** | Follow/unfollow users, recommendations, comment permissions |
| **Tour Service** | Tour creation, management of key points, tour execution sessions |
| **Gateway** | Central API gateway routing requests to appropriate services |

All microservices are isolated and independently deployable. They communicate via REST and RPC where applicable.

---

## User Roles
- **Administrator:** Manage users and monitor activities  
- **Tour Guide:** Create and manage tours  
- **Tourist:** Browse, purchase, and execute tours; leave reviews  

---

## Key Features
1. User registration and role selection (Tourist or Guide)  
2. Admin can view and block users  
3. Profile management (name, bio, profile picture, motto)  
4. Blog creation with markdown support  
5. Commenting and liking functionality for blogs  
6. Following other users and personalized recommendations  
7. Tour creation, draft, published, and archived states  
8. Adding key points to tours with geolocation and media  
9. Tour reviews and ratings by tourists  
10. Tour map visualization and editing  
11. Tour purchase with shopping cart and tokens  
12. Tour execution tracking with position simulation and session management  

---

## Running the Project
1. Clone the repository:
```bash
git clone <repo-url>
cd tourist-application
```
2. Build all Docker containers:
```bash
docker-compose build
```
3. Start the application:
```bash
docker-compose up
```
4. Start monitoring tools:
```bash
cd monitoring
docker-compose up
```

---

## Docker
- Each microservice has its own Dockerfile  
- Docker Compose orchestrates all services and databases  

### Monitoring Stack
The application uses a full monitoring stack to collect and visualize system and container metrics, logs, and traces:  
- **Telegraf** – collects container and host metrics (CPU, RAM, filesystem, network traffic)  
- **Prometheus** – stores metrics and provides query capabilities  
- **Grafana** – dashboards for visualizing metrics  
- **Jaeger** – distributed tracing for tour service requests 
- **Promtail** – collects and ships logs to the logging backend  
- **Node Exporter** – exposes host system metrics to Prometheus  
- **Renderer** – generates visualizations from monitoring data
  
---
