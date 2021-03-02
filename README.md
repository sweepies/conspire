# ![logo](https://user-images.githubusercontent.com/7191851/105656063-3f889680-5e76-11eb-857e-38fab7106630.png) conspire
Conspire is a file sharing server written in Go. It uses any S3-compatible storage as a backend and comes with multi-domain support out of the box. It has the ability to serve a hostname-specific index page and favicon.

## Index
The logic for determining which file to serve as the index is basically the following psuedocode:

```glob("static/index/{host}.*")[0]```


That is, the first file found in the `static/index` directory named with the request hostname (extension ignored) will be served as the index page for that hostname. The extension of that file will be used to determine the MIME type.

## Favicon
Conspire will serve `{host}.ico` under `static/favicon` on the `/favicon.ico` path.

## Caching
Currently, there is no caching implemented. I recommend configuring this at the web server level or using a service like Cloudflare.

## Configuration
TODO

<details>
<summary>AWS credentials are required on top of these values, but are satisfied using the AWS SDK default credential provider chain</summary>

![screenshot](https://user-images.githubusercontent.com/7191851/105654757-86c15800-5e73-11eb-9537-d4832f1c1c65.png)
</details>

| key | required | default | description
| --- | --- | --- | ---
| s3_endpoint | no | s3.amazonaws.com | S3-compatible API endpoint
| s3_region | no | us-east-1 | S3-compatible API region
| s3_bucket | yes | N/A | S3-compatible API bucket
| public_fetch_url | no | N/A | If provided, files are fetched from this URL instead of the S3-compatible API
| set_public_acl | no | false | Whether to set public read access on uploaded objects (most likely for use with public_fetch_url)
| default_cache_control | no | `public, max-age=31536000` | The default Cache-Control value to use when uploading and fetching objects

### Users
Uploading requires HTML basic authentication. Users are configured via `users.json` in the working directory. The schema is as follows:
```json
[
    {
        "username": "sweepyoface",
        "password": "password"
    }
]
```
### Image attribution
<sub>Icons made by [iconixar](https://www.flaticon.com/authors/iconixar) from www.flaticon.com</sub>