# GoDynamicAgentModel

## Motivation
- Playing around with using Golang for ABM

## Summary
- Inspired by [this article]() and [this repo]().
- Mostly straightforward porting of Julia code into Golang

## Notes
- Used `gizak/termui/v3` to provide a terminal UI for plots
- Used a _set_ (okay, map-as-set) of neighbors instead of a sorted list
- Used `gonum/stat/distuv` for the binomial distribution

## Conclusion
- It works
- It's _fast_ :-)
