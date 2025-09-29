package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/chromedp/chromedp"
)

type Product struct {
	URL         string  `json:"url"`
	Name        string  `json:"name"`
	Image       string  `json:"image"`
	Price       string  `json:"price"`
	RatingAvg   float64 `json:"rating_avg"`
	RatingCount int     `json:"rating_count"`
}

const (
	maxProducts = 100
	urlTemplate = "https://raidlight.com/collections/all?page=%d"
)

func main() {
	// Option flag: print text or JSON output
	outputJSON := flag.Bool("json", true, "Output results as JSON")
	flag.Parse()

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Scraping logic
	products, err := scrapeProducts(ctx, maxProducts)
	if err != nil {
		log.Fatalf("Failed to scrape products: %v", err)
	}

	outputResults(products, *outputJSON)
}

// scrapeProducts handles the main scraping logic
func scrapeProducts(ctx context.Context, maxProducts int) ([]Product, error) {
	var products []Product
	page := 1
	seen := make(map[string]struct{})

	for len(products) < maxProducts {
		// Scrape each page
		pageProducts, err := extractProductsFromPage(ctx, page)
		if err != nil {
			return nil, fmt.Errorf("error scraping page %d: %w", page, err)
		}

		if len(pageProducts) == 0 {
			log.Printf("No products extracted from page %d, might be layout change or end of products", page)
			break
		}

		productsAdded := addProductsWithoutDuplicates(&products, pageProducts, seen, maxProducts)

		// If we got no new products from this page, we've likely reached the end
		if productsAdded == 0 && len(products) > 0 {
			fmt.Printf("No new products found on page %d, stopping\n", page)
			break
		}

		page++
	}

	return products, nil
}

// extractProductsFromPage handles page-level scraping and JavaScript execution
func extractProductsFromPage(ctx context.Context, page int) ([]Product, error) {
	var pageProducts []Product
	url := fmt.Sprintf(urlTemplate, page)

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.WaitVisible(".grid-product", chromedp.ByQuery),
		chromedp.Evaluate(extractJSContent, &pageProducts),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to navigate or evaluate page %d: %w", page, err)
	}

	return pageProducts, nil
}

func addProductsWithoutDuplicates(products *[]Product, pageProducts []Product, seen map[string]struct{}, maxProducts int) int {
	addedCount := 0
	for _, product := range pageProducts {
		if len(*products) >= maxProducts {
			break
		}
		if _, exists := seen[product.URL]; !exists {
			seen[product.URL] = struct{}{}
			*products = append(*products, product)
			addedCount++
		}
	}
	return addedCount
}

// outputResults handles the output formatting
func outputResults(products []Product, outputJSON bool) {
	if outputJSON {
		jsonData, err := json.MarshalIndent(products, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(jsonData))
	} else {
		fmt.Printf("These are the %d products Found :\n\n", len(products))
		for i, product := range products {
			if i >= maxProducts {
				break
			}
			fmt.Printf("%d. %s\n", i+1, product.Name)
			fmt.Printf("   URL: %s\n", product.URL)
			fmt.Printf("   Price: %s\n", product.Price)
			if product.RatingCount > 0 {
				fmt.Printf("   Rating: %.2f/5 (%d reviews)\n", product.RatingAvg, product.RatingCount)
			}
			fmt.Printf("   Image: %s\n\n", product.Image)
		}
	}
}

const extractJSContent = `
Array.from(document.querySelectorAll('.grid-product')).map(product => {
  const link = product.querySelector('a.grid-product__link');
  const nameEl = product.querySelector('.grid-product__title');
  const imageEl = product.querySelector('.grid__image-ratio, img');
  const priceEl = product.querySelector('.grid-product__price .money');
  const ratingEl = product.querySelector('.jdgm-prev-badge');

  // Extract rating info
  let ratingAvg = 0;
  let ratingCount = 0;
  if (ratingEl) {
    const avgAttr = ratingEl.getAttribute('data-average-rating');
    const countAttr = ratingEl.getAttribute('data-number-of-reviews');
    ratingAvg = avgAttr ? parseFloat(avgAttr) : 0;
    ratingCount = countAttr ? parseInt(countAttr) : 0;
  }

  // Extract image URL
  let imageUrl = '';
  if (imageEl.tagName === 'IMG') {
	imageUrl = imageEl.src; 
  }

  return {
    url: link ? link.href : '',
    name: nameEl ? nameEl.textContent.trim() : '',
    image: imageUrl,
    price: priceEl ? priceEl.textContent.trim() : '',
    rating_avg: ratingAvg,
    rating_count: ratingCount
  };
}).filter(p => p.url && p.url.includes('/products/'));`
