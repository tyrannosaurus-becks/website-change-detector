# website-change-detector

A tool using Twilio that sends you a text message if a phrase is removed from a web page.

Usage:
```
go build && go install
export TWILIO_ACCOUNT_SID=<account-sid>
export TWILIO_AUTH_TOKEN=<auth-token>
website-change-detector \
    -url='https://google.com' \
    -phrase="I'm Feeling Lucky" \
    -to='503-123-4567' \
    -from='503-098-7654' \
    -frequency=60 \
    -dryrun=true
```

- The `from` phone number must be one that has been created in Twilio.
- The frequency is how often, in seconds, you'd like the given URL checked.
- When dryrun is set to true, no text messages are sent, but log lines indicate
when one would have been set.

Please note that while a `phrase` may appear unbroken on a website (like "Hello, World!"),
in HTML itself, it may be reflected as "<em>Hello,</em> World!". So, it's possible to _see_
words on a page and not have this script find a match. That's why it's important to either
view the target url's HTML when choosing your `phrase`, or ensure it's working the way
you expect by running with the `dryrun` flag at first.
