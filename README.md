# GoKubeBalancer

GoKubeBalancer is a load balancer application written in Go that distributes TCP connections among backend servers. It is designed to provide efficient load distribution and failover capabilities for network applications. The application includes a metrics server for monitoring performance and health endpoints for status checks. It is designed to be deployed to integrate with Rancher to manage backend server configurations dynamically.

## Features

- TCP load balancing for incoming client connections (Limited to port 80 and 443)
- Backend server management for routing client connections
- Monitoring of metrics and health endpoints for performance tracking
- Rancher integration for dynamic backend server configuration

## Installation

To run the GoKubeBalancer application, you need Go installed on your system:
bash
go run main.go

## Configuration

The configuration for GoKubeBalancer is defined in the config/config.yaml file. You can adjust settings such as frontend and backend ports, backend server IPs, and metrics server port in this configuration file.
Usage
Start the metrics server to monitor load balancer performance:
bash
go run main.go metrics
Start the TCP load balancer to distribute client connections:
bash
go run main.go balancer
Check the health and readiness endpoints for status information:
Health Endpoint: <http://localhost:8080/healthz>
Readiness Endpoint: <http://localhost:8080/readyz>

## Docker Build and Push

The project includes a GitHub Actions workflow for building and pushing the Docker image to DockerHub. On each push to the main branch, the workflow builds the image and tags it with version information before pushing it to DockerHub.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Contributing

Contributions are welcome! Please follow the guidelines in the CONTRIBUTING.md file.

## Contact

For any questions or feedback, feel free to contact the project maintainer:
Name: [Your Name]
Email: [Your Email]
