# IP Geolocation Service - Architecture Diagram

## System Architecture

```mermaid
graph TB
    subgraph "External"
        Client[Client Application]
        LB[Load Balancer]
    end
    
    subgraph "API Layer"
        API[API Server :8080]
    end
    
    subgraph "Middleware"
        MW[Middleware Stack<br/>Recovery • Logging • Rate Limit • CORS]
    end
    
    subgraph "Handlers"
        HANDLER[HTTP Handlers<br/>IP • Health • Debug]
    end
    
    subgraph "Services"
        SERVICE[IP Service<br/>+ Validator]
    end
    
    subgraph "Repository"
        REPO[Repository Interface<br/>+ File Implementation]
    end
    
    subgraph "Data"
        DATA[CSV Data File]
    end
    
    Client --> LB
    LB --> API
    API --> MW
    MW --> HANDLER
    HANDLER --> SERVICE
    SERVICE --> REPO
    REPO --> DATA
    
    CONFIG[Configuration] -.-> API
    CONFIG -.-> MW
    CONFIG -.-> REPO
    
    classDef external fill:#e3f2fd,stroke:#1976d2,stroke-width:2px
    classDef api fill:#f3e5f5,stroke:#7b1fa2,stroke-width:2px
    classDef internal fill:#e8f5e8,stroke:#388e3c,stroke-width:2px
    classDef data fill:#fff3e0,stroke:#f57c00,stroke-width:2px
    classDef config fill:#e0f2f1,stroke:#00695c,stroke-width:2px
    
    class Client,LB external
    class API api
    class MW,HANDLER,SERVICE,REPO internal
    class DATA data
    class CONFIG config
```

## Layer Architecture

```mermaid
graph TB
    subgraph "🎨 Presentation Layer"
        A[HTTP Handlers<br/>Middleware Stack<br/>Router]
    end
    
    subgraph "🧠 Business Logic Layer"
        B[IP Service<br/>IP Validator<br/>Business Rules]
    end
    
    subgraph "💾 Data Access Layer"
        C[Repository Interface<br/>File Repository<br/>Data Abstraction]
    end
    
    subgraph "🗄️ Data Layer"
        D[CSV Data File<br/>IP Location Data]
    end
    
    A -->|"HTTP Requests"| B
    B -->|"Data Queries"| C
    C -->|"File Access"| D
    
    classDef presentation fill:#e3f2fd,stroke:#1976d2,stroke-width:3px
    classDef business fill:#f3e5f5,stroke:#7b1fa2,stroke-width:3px
    classDef data fill:#e8f5e8,stroke:#388e3c,stroke-width:3px
    classDef storage fill:#fff3e0,stroke:#f57c00,stroke-width:3px
    
    class A presentation
    class B business
    class C data
    class D storage
```

## Request Flow Sequence

```mermaid
sequenceDiagram
    participant C as 🌐 Client
    participant A as 🚀 API Server
    participant M as 🛡️ Middleware
    participant H as 📝 Handler
    participant S as ⚙️ Service
    participant R as 💾 Repository
    
    Note over C,R: IP Geolocation Request Flow
    
    C->>+A: GET /v1/find-country?ip=1.2.3.4
    A->>+M: Apply Middleware Stack
    Note over M: Rate Limiting • Logging • CORS • Security
    M->>+H: Forward Request
    H->>+S: FindLocation(ip)
    Note over S: Validate IP Address
    S->>+R: FindLocation(ip)
    Note over R: Lookup in CSV Data
    R-->>-S: Return Location Data
    S-->>-H: Return Location
    H-->>-M: Return JSON Response
    M-->>-A: Return Response
    A-->>-C: HTTP 200 + Location JSON
    
    Note over C,R: Request Completed Successfully
```

## Key Features

| Feature | Description |
|---------|-------------|
| 🏗️ **Clean Architecture** | Clear separation of concerns with layered design |
| 🚦 **Rate Limiting** | Token bucket implementation with configurable limits |
| 🏥 **Health Checks** | Service and repository health monitoring |
| 🛡️ **Middleware Stack** | Recovery, Logging, CORS, Security headers |
| 💾 **Repository Pattern** | Data access abstraction with interface-based design |
| 🔧 **Dependency Injection** | Constructor-based DI for testability |
| 📊 **Structured Logging** | JSON/text logging with configurable levels |
| 🐳 **Docker Support** | Multi-stage build with Docker Compose |
