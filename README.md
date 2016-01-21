# a tool for diagnosing MySQL connection errors

This is a tool to help diagnose errors when connecting to a remote MySQL database.

## example

invocation:

    ./mysql-remote-diag-linux \
        -hmysql.example.com \
        -ureporting-user \
        -pso-sekret \
        --ssl-ca /etc/pki/tls/certs/ec2-db.pem \
        --ssl-verify-server-cert \
        some-db

result, showing an invalid TLS certificate on the server:

    {
        "Version": "¯\\_(ツ)_/¯",
        "Options": {
            "ShowHelp": false,
            "Debug": false,
            "ExternalIPLookupURL": "http://ifconfig.co/",
            "Host": "mysql.example.com",
            "Port": 3306,
            "User": "reporting-user",
            "Password": "so-sekret",
            "SSLCA": "/etc/pki/tls/certs/ec2-db.pem",
            "SSLVerify": true,
            "Args": {
                "DBName": "some-db"
            }
        },
        "DBName": "mysql",
        "DSN": "reporting-user:so-sekret@tcp(10.11.12.13:3306)/mysql?timeout=1m\u0026tls=pre-resolved",
        "ExternalIP": "127.0.0.42",
        "MySQLServerAddr": {
            "IP": "10.11.12.13",
            "Zone": ""
        },
        "Passed": false,
        "Result": "error connecting: x509: certificate is valid for servercert, not mysql.example.com"
    }

## building

Just run `make`.  Binaries for your host platform (tested on OS X), linux amd64, and windows 386 are written to the `stage` directory.
