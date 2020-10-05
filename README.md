# GoDynamicAgentModel

## Motivation
- Playing around with using Golang for ABM

## Summary
- Inspired by [this article]() and [this repo]().
- Mostly straightforward porting of Julia code into Golang

## Sample invocation

```shell
$ go run main.go  --v=3 --log_dir=/home/agam/tmp/
```

## Notes
- Used `gizak/termui/v3` to provide a terminal UI for plots
- Used a _set_ (okay, map-as-set) of neighbors instead of a sorted list
- Used `gonum/stat/distuv` for the binomial distribution

## Conclusion
- It works
- It's _fast_ :-)

## Sample screenshot
[Screenshot](https://github.com/agam/GoDynamicAgentModel/edit/master/SampleScreenShot.png)
