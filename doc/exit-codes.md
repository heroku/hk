# Exit Codes

The termination status of hk is communicated by exit code. Where possible, hk tries to follow existing convention around exit codes, and introduces custom exit codes in the range of 64 - 113 as suggested by [the Advanced Bash-scripting Guide](http://tldp.org/LDP/abs/html/exitcodes.html).

The following is a list of exit codes currently used by hk:

| Exit Code | Description                                |
|:---------:| ------------------------------------------ |
| 0         | Success                                    |
| 2         | Command usage error                        |
| 79        | Second authentication factor required      |
