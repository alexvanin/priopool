<p align="center">
<img src="/.github/logo.svg" width="500px">
</p>
<p align="center">
  Goroutines pool with priority queue buffer.
</p>

---

## Overview

Package `priopool` provides goroutines pool based on
[panjf2000/ants](https://github.com/panjf2000/ants) library with priority queue 
buffer based on [stdlib heap](https://pkg.go.dev/container/heap) package.

Priority pool:
- is non-blocking,
- prioritizes tasks with higher priority value,
- can be configured with unlimited queue buffer.

## Install

```powershell
go get -u github.com/alexvanin/priopool
```

## License

Source code is available under the [MIT License](/LICENSE).
