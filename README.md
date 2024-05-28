### GophKeeper Client

GophKeeper is a client-server system designed to securely store and manage user credentials, binary data, and other private information. This repository contains the client-side implementation of GophKeeper, providing a command-line interface (CLI) for users to interact with the GophKeeper server.

#### Key Features

- **User Authentication and Authorization**: Allows users to authenticate and authorize against a remote GophKeeper server.
- **Data Access**: Provides secure access to stored private data on the server upon request.
- **Platform Compatibility**: The client is compatible with Windows, Linux, and Mac OS platforms.
- **Version Information**: Users can retrieve the version and build date of the client binary.

#### Supported Data Types

- Login/Password pairs
- Arbitrary text data
- Arbitrary binary data
- Credit card information

All data types can include additional textual metadata, such as associated websites, personal identities, banks, and lists of one-time activation codes.

#### Project Structure

- **cmd/gophkeeper**: Contains the main CLI application entry point.
  - `main.go`: Entry point for the client application.
  - `server.crt`, `server.key`: Certificates for secure communication (placeholder).
  - `session.dat`, `syncinfo.dat`: Data files for session and synchronization information.
  
- **pkg**: Contains various packages used by the client.
  - **appcontext**: Application context management.
  - **bdkeeper**: Database management and migrations.
  - **client**: Client communication logic.
  - **config**: Configuration management.
  - **encription**: Encryption utilities.
  - **gksync**: Synchronization logic.
  - **logger**: Logging utilities.
  - **models**: Data models.
  - **services**: Core services.
  - **syncinfo**: Synchronization information management.
  
- **api-spec**: API specifications for client-server interactions.
  - `gophkeeper-api-spec.yaml`: Current API specification.
  - `gophkeeper-api-spec-old.yaml`: Previous version of the API specification.

- **.github/workflows**: GitHub Actions workflows for CI/CD.

#### Getting Started

1. **Installation**: Download the appropriate binary for your platform from the releases page.
2. **Configuration**: Configure the client using the provided configuration file template.
3. **Usage**:
   - Register a new user or authenticate an existing user.
   - Add or request private data.
   - Data is synchronized with the GophKeeper server.

#### Building from Source

1. Clone the repository:
   ```sh
   git clone https://github.com/yourusername/gophkeeper-client.git
   cd gophkeeper-client
   ```
2. Build the client:
   ```sh
   go build -o gophkeeper cmd/gophkeeper/main.go
   ```

#### Testing

The client is extensively tested with unit tests to ensure reliability and stability. Run the tests using:
```sh
go test ./...
```

#### Contribution

Contributions are welcome! Please submit pull requests or open issues to discuss potential improvements or report bugs.

#### License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

For more detailed documentation, please refer to the [README](README.md) file in the repository.
 