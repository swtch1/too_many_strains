FROM mysql:5.7.12

ARG MYSQL_USER=root
ARG MYSQL_PASSWORD=password

ENV MYSQL_ROOT_PASSWORD=${MYSQL_PASSWORD}

ADD *.sql /docker-entrypoint-initdb.d/
COPY my.cnf /etc/mysql/
