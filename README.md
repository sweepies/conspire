# ![logo](https://user-images.githubusercontent.com/7191851/105656063-3f889680-5e76-11eb-857e-38fab7106630.png) conspire
Conspire is a file sharing server written in Go. It uses any S3-compatible storage as a backend and supports caching. It supports any number of domains, and has the ability to serve a hostname-specific index page and favicon.

## Index
The first file found in the `static/index` directory with the name of the request hostname, excluding the extension, will be served as the index page for that hostname. The extension of that file will be used to determine the MIME type.

## Favicon
Conspire will serve `{host}.ico` under `static/favicon` on the `/favicon.ico` path.

## Caching
Currently, caching is controlled only by a single boolean variable. If it is enabled, responses will be cached in memory for 30 minutes (this includes the index page and favicon, to save disk IO).

## Configuration
Conspire uses [viper](https://github.com/spf13/viper) to fetch configuration values. This means it supports environment variables, JSON, TOML, YAML, HCL, envfile and Java properties config files. Any of these should work as long as the file is called `config.{ext}` and in the same directory as the binary.

<details>
<summary>AWS credentials are required on top of these values, but are satisfied using the AWS SDK default credential provider chain</summary>

![screenshot](https://user-images.githubusercontent.com/7191851/105654757-86c15800-5e73-11eb-9537-d4832f1c1c65.png)
</details>

| key | required | default | description
| --- | --- | --- | ---
| s3_endpoint | no | s3.amazonaws.com | S3-compatible API endpoint
| s3_region | yes | us-east-1 | S3-compatible API region
| s3_bucket | yes | N/A | S3-compatible API bucket
| cache_enabled | no | true | Whether or not to cache responses in memory

## TODO
- [ ] Add uploading support
- [ ] Add tests
- [ ] Improve authentication scheme (currently HTTP basic auth)