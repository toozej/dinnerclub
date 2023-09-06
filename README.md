# dinnerclub

 Web application to facilitate reviewing and organizing get-togethers at restaurants

 Uses [toozej/golang-starter](https://github.com/toozej/golang-starter) as a base

## Deployment via Fly.io
### First time setup
#### Postgres Database
- fly postgres create
    - app name: ${CITYCODE}dinnerclub-db
    - single node
    - don't scale to zero after one hour
    - add connection info to deploy.env, specifically DATABASE_URL
    - create dinnerclub user and database
        - `flyctl proxy 5434:5432 -a $(DEPLOY_APPNAME)-db &`
        - `PGPASSWORD=$(DEPLOY_POSTGRES_PASSWORD) psql -h localhost -p 5434 -U postgres`
            - `CREATE ROLE ${CITYCODE}dinnerclub WITH LOGIN PASSWORD 'passwordhere' CREATEDB;`
            - `CREATE DATABASE ${CITYCODE}dinnerclub;`
            - `GRANT ALL PRIVILEGES ON DATABASE ${CITYCODE}dinnerclub TO ${CITYCODE}dinnerclub;`
            - `\c ${CITYCODE}dinnerclub`
            - `GRANT ALL ON SCHEMA public TO ${CITYCODE}dinnerclub;`
            - `\q`
        - `pkill -15 -f 'flyctl proxy'`
