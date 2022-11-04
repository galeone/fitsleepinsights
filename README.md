# Sleepbit

Web application that connects with your [Fitbit](https://www.fitbit.com/) account and tries to do explainable predictions of your sleep quality - using all the available data.

- What makes you sleep well?
- What are the activities you should perform and when?
- What you should avoid?
- What increases your deep sleep?

These are some of the question this project tries to answer.

The goal is to understand what behaviors/activities are good/bad for *your* sleep. People are different, and what makes a person sleep well could make some one else sleep a nightmare. For this reason, the prediction are made *per user* - no data is aggregated. 

## The idea



## Technicalities

The application is made of different parts

- A Fitbit API client. Folder `fitbit`
- A storage. The storage requires a dedicated setup (we use TimescaleDB/PostgreSQL). Folder `storage`.
- The predictive models
- The web application

The web application, together with the `client` and the `database` asks the user permission on the Fitbit account.

### Database setup

1. Install TimescaleDB: https://docs.timescale.com/install/latest/self-hosted/installation-archlinux/

2. Create the user and the database

```bash
sudo -i -u postgres createuser -U postgres sleepbit
sudo -i -u postgres createdb -U postgres sleepbit sleepbit
```

Done. On startup sleepbit creates the schema if not present.


### Running

Yup, running is important.

```bash
go run main.go
```

Or you can install it with `go install` and execute `sleepbit`.

The application can be configured via `.env` file.

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
DB_USER=sleepbit
DB_PASS=sleepbit
DB_NAME=sleepbit

# Web application settings
DOMAIN=sleepbit.com
PORT=8989
```
