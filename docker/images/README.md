On generating self signed certs for testing:

https://www.akadia.com/services/ssh_test_certificate.html

openssl genrsa -des3 -out server.key 4098
openssl req -new -key server.key -out server.csr
mv server.key{,.org}
openssl rsa -in server.key.org -out server.key
openssl x509 -req -days 365 -in server.csr -signkey server.key -out server.crt

