# IP Whitelist Service

This Go program creates a simple web service that allows users to add their IP address to an iptables whitelist by providing a correct password.

## Features

- Web interface showing the user's IP address
- Password-protected IP whitelisting
- Configurable port, password, and iptables chain name
- Adds IP to specified iptables chain upon successful authentication

## Prerequisites

- Go 1.13 or higher
- Root privileges (for modifying iptables rules)
- Linux system with iptables

## Installation

1. Clone this repository:
   ```
   git clone https://github.com/yourusername/ip-whitelist-service.git
   cd ip-whitelist-service
   ```

2. Build the program:
   ```
   go build ip_whitelist_service.go
   ```

## Usage

Run the program with root privileges, specifying the port, password, and iptables chain name:

```
sudo ./ip_whitelist_service -port 8080 -password your_secure_password -chain DOCKER-USER
```

Replace `8080` with your desired port, `your_secure_password` with a secure password, and `DOCKER-USER` with the name of the iptables chain you want to modify.

## Security Considerations

- This program requires root privileges to modify iptables rules. Use with caution.
- In a production environment, implement more secure methods for password storage and verification.
- This implementation does not handle concurrent requests. Consider adding proper synchronization for production use.
- Iptables rules may be reset on system restart. Consider implementing a method to persist these rules.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
