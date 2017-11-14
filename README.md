# Harbormaster

Tool to check azkaban executions

## Usage

1. Authenticate using login. This will request a session id (which is valid for 24h):
```
  $ harbormaster login <azkaban url> <user> <password>
  2017/11/12 12:53:54 Authenticating against https://<azkaban url>...
  export AZKABAN_SESSION_ID=<session id>
  export AZKABAN_HOST=<azkaban url>
```
Paste these lines into your terminal and harbormaster will use these for all subsequent calls.
2. Set up shell completions.

For zsh: `eval "$(harbormaster  --completion-script-zsh)"`
For bash: `eval "$(harbormaster  --completion-script-bash)"`

3. List executions for a flow:

```
$ harbormaster executions <project> <flow>

1676218          RUNNING         Sun, 12 Nov 2017 09:21:37 EST  3h36m4.976537869s        3 hours ago
1676142          FAILED          Sun, 12 Nov 2017 08:49:11 EST  1m45.296s                        4 hours ago
1674649          FAILED          Sat, 11 Nov 2017 22:45:44 EST  9h7m25.781s              14 hours ago
1669533          FAILED          Fri, 10 Nov 2017 11:47:04 EST  4h54m43.673s             2 days ago
1668834          SUCCEEDED       Fri, 10 Nov 2017 07:07:18 EST  4h27m53.965s             2 days ago
1662206          SUCCEEDED       Wed, 08 Nov 2017 10:45:41 EST  7h41m51.985s             4 days ago
```

4. Check status for a flow:

```
$ harbormaster check flow <project> <flow>

Checking status of <project>::<flow>...
Job health:      healthy
Stats:           2 failures, 18 successes, 0 running, 20 total
Last success:    1 hour ago
Histogram:       .XX.................
```

# References

http://azkaban.github.io/azkaban/docs/latest/#ajax-api


# Ideas & TODO

- [ ] add report feature to summarize execution times for a project's flows
- [ ] on check view show if job is currently running, and how long
- [ ] don't capture session expiration gracefully
- [ ] download and cache project structure
- [ ] add tab completion for projects, flows, and executions
