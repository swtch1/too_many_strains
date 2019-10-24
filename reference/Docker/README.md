# Too Many Strains - MySQL Docker

This docker image is provided to help you build this backend challenge

```
docker run --name flourish-mysql --publish 3306:3306 -d --rm flourish-mysql:latest
docker container kill flourish-mysql
```

This stops your running container and deletes all the data within it.
