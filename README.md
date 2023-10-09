# choco-internalize
internalize choco package by nexus

### 
download cache proxy setup
nginx.conf
```
server {
    listen 80;

    location ~* /redirect/https/(.+)$ {
        return 302 https://$1$is_args$args;
    }
    location ~* /redirect/http/(.+)$ {
        return 302 http://$1$is_args$args;
    }
}
```
Setup nexus proxy nginx instance, then you can cache any file on nexus
