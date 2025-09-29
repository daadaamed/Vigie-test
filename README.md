# Product Scraper

A Go script that scrapes product data from raidlight.com using ChromeDP.

## Features

- Scrapes up to 100 products with details (name, price, rating, image)
- Outputs results as JSON or formatted text
- Handles pagination automatically
- Removes duplicate products

## Usage

```bash
# Output as JSON (default)
go run main.go

# Output as formatted text
go run main.go -json=false