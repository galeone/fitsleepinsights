# Fit Sleep Insights

Web application that connects with your [Fitbit](https://www.fitbit.com/) account and tries to do explainable predictions of your sleep quality - using all the available data.

- What makes you sleep well?
- What are the activities you should perform and when?
- What you should avoid?
- What increases your deep sleep?

These are some of the question this project tries to answer.

The goal is to understand what behaviors/activities are good/bad for *your* sleep. People are different, and what makes a person sleep well could make some one else sleep a nightmare. For this reason, the prediction are made *per user* - no data is aggregated. 

## Technicalities

The application is made of different parts

- A Fitbit API client. Folder `fitbit`
- A storage. The storage requires a dedicated setup (we use TimescaleDB/PostgreSQL). Folder `storage`.
- The predictive models
- The web application

The web application, together with the `client` and the `database` asks the user permission on the Fitbit account.

### Docker-Compose

You can chose to configure the application via the `docker-compose.yaml` file, and executed it trough `docker compose` - or you can configure (without depending on Docker) the various components of the application locally.

Below you can find the steps to setup the application locally.

### Database setup

1. Install PostgreSQL

2. Create the user and the database. Note: on MacoOS skip the `sudo -i -u postgres ` part of the commands.

```bash
sudo -i -u postgres createuser -U postgres fitsleepinsights
sudo -i -u postgres createdb -U postgres fitsleepinsights fitsleepinsights
```

3. Grant the create privilege (required from PostgreSQL 15+), and make it superuser (required for the vector extension):

```
psql -U postgres -d fitsleepinsights -c "GRANT USAGE, CREATE ON SCHEMA public TO fitsleepinsights;"
psql -U postgres -d fitsleepinsights -c "ALTER USER fitsleepinsights WITH SUPERUSER;"
```

4. Install pgvector: https://github.com/pgvector/pgvector

Done. On startup the application creates the schema if not present.

### Configuration

Place a `.env` file in the runtime path. This is the content:

```env
# Register the application
# https://dev.fitbit.com/build/reference/web-api/developer-guide/getting-started/
# for obtainign these values.
FITBIT_CLIENT_ID=""
FITBIT_CLIENT_SECRET=""
FITBIT_REDIRECT_URL=""

# Use the same values used in the Database setup section
DB_HOST=localhost
DB_PORT=5432
DB_USER=fitsleepinsights
DB_PASS=fitsleepinsights
DB_NAME=fitsleepinsights


# Web application settings
DOMAIN=fitsleepinsights.app
PORT=8989

# Vertex AI
VAI_LOCATION="europe-west6"
VAI_PROJECT_ID="project id"
VAI_SERVICE_ACCOUNT_KEY="full path"
```


### Running

Yup, running is important. There are 2 ways of running the application, depending on how you configured it.

If you installed the database and configured the `.env` file you need simply to run it.

```bash
go run main.go
```

If you configured the `docker-compose.yaml` file, than you can:

```bash
./run.sh up
# and to stop it
./run.sh down
```

Or you can install it with `go install` and execute `fitsleepinsights`.

