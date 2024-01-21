# Portfoliotools

A repository of tools that help manage a portfolio of assets (equities, bonds, etc.).  
Tools include:
* Current annualized returns (CAR)
* Target annualized returns (TAR)
* Stock/ETF/Ticker lookup
  * calculate volatility for ticker
  * calculate volatility-adjusted range using  
  weighted volume average price (vwap) of stock/ETF

Portfolio tools was created with the express purpose of helping identify opportunities to enter and exit positions in a 
portfolio and identify potential targets to initiate positions that you are already monitoring. This tool is an aid and 
not to be relied upon solely for decision-making. This should be used to augment any process you have for building a 
portfolio. Over time, more tools will be added to this package to enhance the capabilities and analysis possibilities, 
but this should be used at your own risk to enhance your own process.  

CAR (Current Annualized Returns): This tool helps you understand what your CURRENT annualized return is as of the 
current day. It takes your date of position initialization (date of purchase/sale) and calculates what your return 
would be if you were to exit the position with the same rate of return expanded out to a year. So, if you have a goal 
of a 13% Annual Return, you would check the output of this function against that value to determine if you should exit. 
This works for both long and short positions held. You just need to provide a flag to indicate that it is a short 
position along with the other required variables. You need:
* date of purchase (startDate)
* current price
* purchase price (costBasis)

TAR (Target Annualized Returns): This tool helps you understand what price an asset would have to be in order for you to exit your position at your 
targeted rate of return for that specific day. This tool is very similar to the Current Annualized Returns (CAR), except
instead of giving you your current return, it tells you what price you would need to hit your target return. If you 
wanted to target the previous 13%, you would use .13 as the target rate, you can figure out what price you would need 
for an asset to hit your target rate, so you could set a limit order to exit the position based on the target price. You
need:
* purchase price (costBasis)
* purchase date (startDate)
* target rate (targetRate)

The main tool here is the stock lookup tool. It gathers data from the start date to the end date for a stock and uses 
polygon.io to gather data. You're going to need at least a free account with them in order to use this tool function. 
Once you've gotten your API key for polygon.io, you need to set it up in a config file as described in the user guide 
set up section. 

## Table of Contents

- [Overview](#portfoliotools)
- [Table of Contents](#table-of-contents)
- [User Guide](#user-guide)
  - [Setup](#set-up)
  - [Usage](#usage)
    - [CAR - Current Annualized Return](#car)
    - [TAR - Target Annualized Return](#tar)

## User Guide

### Set Up
You should be able to download a binary copy of the tool from github, eventually (WIP). There will eventually be a 
process to build release versions of the code so that you can just download and use on your preferred platform 
(x86 or ARM and Windows, MacOS, or Linux) and then you would just run with `./<executable_name> <flags>` in whatever 
directory you downloaded with on Mac/Linux and you'd just call the .exe from the directory in Windows.

If you want to build from scratch, you'll need to have the go compiler installed, which you can get the latest version 
of from [go.dev](https://go.dev/doc/install) and follow their install instructions. Once you have go installed, you're 
 going to want to get the sourcecode from this repository:

`git clone https://github.com/khrystoph/portfoliotools.git`

Then change directories into the directory where this was stored (by default, this will be `portfoliotools`). Once in 
the base github directory, you're going to build the project like this:

`go build -o stockclient cmd/stockClient.go`

which should fetch all your dependencies and build your package.

### Usage
Basic usage of this tool:

`./stockclient [flags] [TAR,CAR]`

There are three base functions of this tool:
1. CAR - Current Annualized Return
   * Takes the cost basis, current price, and purchase date of a position (short or long) and calculates the annualized
   return.
2. TAR - Target Annualized Return
   * Takes the cost basis, purchase date, and a target rate to return the price you should sell at today to get the 
   targeted return rate.
3. No argument given
   * Default operation. It pulls a stock price given a start date, end date and stock/ETF ticker and returns various 
   properties about the stock (30/60/90 day volatility, open/close/high/low price, vwap, volume, etc.)

For all three of these different functions, you will have different flags that are required. Many flags are not 
required for different functions of the same tool. The required flags will be discussed in the respective sections 
below.

#### CAR
CAR, also known as Current Annualized Return is an argument supplied to the stock client to help track the current 
annualized return of specific positions in profit (you could track positions in loss, but the absolute return is more 
meaningful as closing out the position would lock in the loss as an absolute value). It helps make decisions on 
when you should trim or close out a position entirely, such as if it hits your minimal threshold targets for total 
portfolio returns for a year. So, if I have a target return rate on my portfolio of 15%, I might want to trim a 
position when the current annualized rate hits 15%, so that I can re-allocate that capital to another position in order 
to compound returns in the portfolio and keep capital actively deployed.

To use the CAR argument, you need Three things:
1. cost basis
2. current price
3. purchase date
4. short (optional for if you are short)

It will use the current date and time to calculate what your current 
annualized return is on an asset in your portfolio. It will look something like this:

```
./stockclient -startTime "2023-05-22T00:00:00Z" -costBasis 17.85 -currentPrice 23.60 CAR
```

The output would look something like this:
```
Current Annualized Returns Selected
Hours owned: 5832.000000.
Current Annualized return is: 0.521109.
```

This means that your current annualized return is +52.1109%. The formatting of the "startTime" is important because of 
how go formats dates and times and how I chose to handle date and time strings. The "Z" means "Zulu" timezone (UTC), 
but you can enter any timezone by changing to it's offset format, which looks something like this:

```2023-05-22T00:00:00-05:00```

as the same date and time, but a different timezone offset (EST). In the case of Eastern Daylight-savings Time, you 
would instead use -04:00. However, most of this is moot as timezones are currently dropped in the current code to 
simplify the code a bit and since we consider "time owned" to be that same day from the beginning of the day. On short 
timeframes, this will have some impact on the annualized rate, but it will result in a lower-than-reality rate, meaning 
your return is actually a bit HIGHER than what the output would be if tracking exact times. However, since granularity 
on purchase/sale is only after the close of business for a day, it doesn't have a material impact, overall.

#### TAR
TAR, or Target Annualized Return is a tool that lets you check an asset in your portfolio to see what price you would 
need to sell it (or buy to cover a short position) in order to get a specific return rate. It takes the purchase date 
(startTime), purchase price (cost basis), target rate (targetRate), 