## Appdiff

[![Build Status](https://travis-ci.org/bl-core-vitals/appdiff.svg?branch=master)](https://travis-ci.org/bl-core-vitals/appdiff)

A CLI tools to compare the difference between two APKs or IPAs.

Useful to reduce manual effort to make a report size app when a new version was released.

## Usage
1. Get `appdiff` executable file from repo 
2. 
```
$ appdiff <new_apk_or_ipa> <old_apk_or_ipa> <dir_level> <custom_outputs_directory>
```
or, use file develop script
```
$ go run main.go  <new_apk_or_ipa> <old_apk_or_ipa> <dir_level> <custom_outputs_directory>
```
until you see
```
...
All data has been copied to clipboard!
$ _
```
> `<dir_level>` and `<custom_outputs_directory>` are optional,
> but `<dir_level>` will become mandatory if you pass `<custom_outputs_directory>` as well

3. Open create a new sheet page to paste the result
4. To be a pretty view, just go to `Data > Split text to columns` 
5. Go ahead to sort by `diff` header or other columns

## Building the Executables

you can use this command to build binary via mac

```sh
go build -o mac-appdiff # or
make build
```

for cross-platform (build linux executable via mac)

```sh
GOARCH="amd64" GOOS="linux" go build -o linux-appdiff # or
make build-linux
```

## Todo
- [ ] Diff by packages in Android
- [ ] Beautiful report 
 
## License

MIT
