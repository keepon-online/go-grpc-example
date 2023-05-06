
openssl ecparam -genkey -name secp384r1 -out server.key

openssl req -nodes -new -x509 -sha256 -days 3650 -config server.cnf -extensions 'req_ext' -key server.key -out server.crt

