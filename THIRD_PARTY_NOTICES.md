# Third-Party Notices

`goskia` is a CGo binding. It builds on, and redistributes artifacts from, the
following third-party projects.

## Skia

- Website: https://skia.org
- Copyright (c) Google LLC
- License: BSD-3-Clause (https://github.com/google/skia/blob/main/LICENSE)

The C-API headers vendored under `skia/csrc/include/c/` originate from Skia
(via the SkiaSharp fork, see below).

## SkiaSharp

- Repository: https://github.com/mono/SkiaSharp
- Copyright (c) .NET Foundation and Contributors; Microsoft Corporation
- License: MIT (https://github.com/mono/SkiaSharp/blob/main/LICENSE.md)

`goskia` links against the prebuilt `libSkiaSharp` native libraries published
by the SkiaSharp project on NuGet, and vendors the project's `include/c/*.h`
C-API headers. These provide a stable C ABI over Skia's C++ API.

- Pinned SkiaSharp version: `3.119.4`
- Pinned `mono/skia` fork commit (headers): `7dbfc07dd33181f84e0958afb7ee805c6c769f0b`

The native binaries are downloaded at setup time from NuGet; they are not
re-distributed inside this repository.
