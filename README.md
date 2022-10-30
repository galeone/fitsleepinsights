# Sleepbit


## Database setup

1. Install TimescalDB: https://docs.timescale.com/install/latest/self-hosted/installation-archlinux/

2. Create the user and the database

```bash
sudo -i -u postgres createuser -U postgres sleepbit
sudo -i -u postgres createdb -U postgres sleepbit sleepbit
```

Done. On startup sleepbit creates the schema if not present.
