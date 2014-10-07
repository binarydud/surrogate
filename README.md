Configuration Sample
--------------------

```toml
[fontends]
[frontends.test]
bind="0.0.0.0:8080"
backends=["server1"]
strategy="round robin"

[backends]
[backends.server1]
hosts=["http://127.0.0.1:8000"]
path="/"


```
