## Appdiff

[![Build Status](https://travis-ci.org/bl-core-vitals/appdiff.svg?branch=master)](https://travis-ci.org/bl-core-vitals/appdiff)

A CLI tools to compare the difference between two APKs or IPAs.

Useful to reduce manual effort to make a report size app when a new version was released.

## Usage
1. Get `appdiff` file from repo 
2. 
```
$ appdiff <new_apk_or_ipa> <old_apk_or_ipa>
```
or, use file develop script
```
$ go run main.go  <new_apk_or_ipa> <old_apk_or_ipa>
```
until you see
```
...
All data has been copied to clipboard!
$ _
```
3. Open create a new sheet page to paste the result
4. To be a pretty view, just go to `Data > Split text to columns` 
5. Go ahead to sort by `diff` header or other columns

## Todo
- [ ] Diff by packages in Android
- [ ] Beautiful report 
 
## License

MIT
