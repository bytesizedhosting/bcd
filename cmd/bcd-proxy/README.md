## Bcd-proxy

Bcd-proxy is a add-on module for Bytesized Connect. It provides BCD with a system for 'pretty URLs' to the various applications that BCD allows you to install. Every BCD user gets a free domain in the form of username.bysh.io.

### Custom domains
You can configure and bring your own domains via the BCD web-interface on bytesized-hosting.com. Besides vanity this will also enable the use of https

### HTTP & HTTPS
Bcd-proxy runs on port 80 and 443 by default and can work with both https and http URLs. Whenever a https URL is given it will attempt to automatically get a certificate for the given domain and use that. Because of let's encrypt's rate limiting https is only supported for custom domains.

### Proxying to a local webserver
You can use the `--unknown-host` flag if you want to use bcd-proxy in combination with a local webserver. Any URLs that it doesn't hold a record for will be passed to the local webserver. Do note that https URLs are also supported but that the communication from bcd-proxy to your local webserver will always run over http. Since the requests are usually local this however should not have security implications.

It's recommended to edit your upstart/systemd files and add the option in so it survives restarts.
