# Binary Protocol

The server speaks with a client a request/response protocol over a persistent TCP connection

## Byte Order

All multi-byte integer fields use big-endian byte order.

## Request Frame

Each request has a 21-byte header followed by an optional body. The first three field are a fixed-size header:

| Offset |     Size | Field          |
| -----: | -------: | -------------- |
|    `0` |      `1` | Operation code |
|    `1` |     `16` | Key            |
|   `17` |      `4` | Body size      |
|   `21` | variable | Body           |

The body size is an unsigned 32-bit value. A zero body size means no body bytes follow.

## Response Frame

Each response has a 5-byte header followed by an optional body. The first 2 fields are a fixed-size header:

| Offset |     Size | Field       |
| -----: | -------: | ----------- |
|    `0` |      `1` | Status code |
|    `1` |      `4` | Body size   |
|    `5` | variable | Body        |

## Public Operations

| Code | Name      | Body                    |
| ---: | --------- | ----------------------- |
|  `0` | `Set`     | Value bytes             |
|  `1` | `Get`     | None                    |
|  `2` | `Del`     | None                    |
|  `3` | `Exist`   | None                    |
|  `4` | `TTL_Set` | 4-byte unsigned seconds |
|  `5` | `TTL_Get` | None                    |
|  `6` | `TTL_Del` | None                    |

`TTL_Expire` is an internal Raft operation and is not part of the public client API.

## Status Codes

| Code | Name               | Meaning                                   |
| ---: | ------------------ | ----------------------------------------- |
|  `0` | `OK`               | Operation succeeded                       |
|  `1` | `KeyAlreadyExists` | Operation rejected because the key exists |
|  `2` | `NoSuchKey`        | Key does not exist or has expired         |
|  `3` | `NoTTL`            | Key has no TTL                            |
|  `4` | `WrongInput`       | Invalid operation body or argument        |
|  `5` | `InternalError`    | Server-side failure                       |
|  `6` | `DeadlineExceeded` | Operation exceeded its deadline           |
|  `7` | `NotLeader`        | Mutation reached a Raft follower          |

For `NotLeader`, the response body may contain the current leader transport address as UTF-8 text.
