#!/usr/bin/env python3
from datetime import datetime
import argparse
import logging


def main():
    parser = argparse.ArgumentParser(description="This is a tool to calculate minimal annualized returns based on"
                                                 "length of asset owned, asset price cost basis, and the risk-free"
                                                 " rate")
    parser.add_argument('-d', '--purchase_date', type=lambda s: datetime.strptime(s, '%Y-%m-%d'),
                        help='Enter date in YYYY-MM-DD format')
    parser.add_argument('-r', '--risk_free_rate', type=float, help="Enter the risk-free rate to calculate against in "
                                                                   "floating point format. Ex.: .03")
    parser.add_argument('--cost_basis', '-c', type=float, help="Enter cost basis of asset")
    args = parser.parse_args()

    return_price = (((datetime.now() - args.purchase_date).days * (args.risk_free_rate / 365.0) + 1)
                    * args.cost_basis)

    print(f"Price of asset needed to close position/take profit: {return_price:.2f}\n")


if __name__ == '__main__':
	main()
