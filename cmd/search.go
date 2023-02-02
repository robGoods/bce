package cmd

import (
	"fmt"
	"github.com/robGoods/bce/pkg"
	"github.com/spf13/cobra"
	"math"
)

var (
	page        int
	rows        int
	asset       string
	fiat        string
	transAmount int
	barkId      string
	payTypes    []string
)

func init() {
	SearchCmd.PersistentFlags().IntVarP(&page, "page", "p", 1, "page list for search page num?")
	SearchCmd.PersistentFlags().IntVarP(&rows, "rows", "r", 10, "rows list for search adv rows list?")
	SearchCmd.PersistentFlags().StringVarP(&asset, "asset", "a", "USDT", "asset for searchKey?")
	SearchCmd.PersistentFlags().StringVarP(&fiat, "fiat", "f", "CNY", "fiat for searchKey?")
	SearchCmd.PersistentFlags().StringVarP(&barkId, "barkId", "b", "", "bark app push message to iphone for searchKey?")
	SearchCmd.PersistentFlags().StringSliceVarP(&payTypes, "payTypes", "y", []string{"BANK", "ALIPAY", "WECHAT"}, "payTypes for searchKey?")
	SearchCmd.PersistentFlags().IntVarP(&transAmount, "transAmount", "t", 0, "transAmount for searchKey?")
	rootCmd.AddCommand(SearchCmd)
}

var SearchCmd = &cobra.Command{
	Use:   "search",
	Short: "search c2c trade for binance.com",
	Long:  `A Fast search c2c trade for binance.com cron for binance.com`,
	Run: func(cmd *cobra.Command, args []string) {
		advSell, err := pkg.SearchAdv("SELL", "USDT", "CNY", page, rows, transAmount, payTypes)
		if err != nil {
			fmt.Println(err)
			return
		}

		var maxSellPrice float64
		var minBuyPrice float64 = math.MaxInt
		for _, adv := range advSell {
			if adv.Adv.Price > maxSellPrice {
				maxSellPrice = adv.Adv.Price
			}
		}

		advBuy, err := pkg.SearchAdv("BUY", "USDT", "CNY", page, rows, transAmount, payTypes)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, adv := range advBuy {
			if adv.Adv.Price < minBuyPrice {
				minBuyPrice = adv.Adv.Price
			}

			if adv.Adv.Price < maxSellPrice {

			}
		}

		fmt.Printf("minBuyPrice : %.2f, maxSellPrice %.2f, BuyTotal %d, SellTotal %d \n", minBuyPrice, maxSellPrice, len(advBuy), len(advSell))
		//if maxSellPrice > minBuyPrice {
		for _, adv := range advBuy {
			if adv.Adv.Price < maxSellPrice {
				var cAdvSell = make([]pkg.Adver, 0)
				for _, adv1 := range advSell {
					//if adv1.Adv.Price > adv.Adv.Price && adv.Adv.MinSingleTransAmount < adv1.Adv.DynamicMaxSingleTransAmount {
					if adv1.Adv.Price > adv.Adv.Price {
						cAdvSell = append(cAdvSell, adv1)
					}
				}

				if len(cAdvSell) > 0 {
					fmt.Println("++++++++++++++++++++++++++++++++++++++++++++")
					fmt.Printf("→ Buy Adv: %s,\tPrice: %.2f, TransAmount  %.2f - %.2f\n", adv.Advertiser.NickName, adv.Adv.Price, adv.Adv.MinSingleTransAmount, adv.Adv.DynamicMaxSingleTransAmount)
					fmt.Println("→")
					msg := fmt.Sprintf("→ Buy Adv: %s,\tPrice: %.2f, Amount  %.2f - %.2f\n", adv.Advertiser.NickName, adv.Adv.Price, adv.Adv.MinSingleTransAmount, adv.Adv.DynamicMaxSingleTransAmount)
					for _, v := range cAdvSell {
						fmt.Printf("← Sell Adv: %s \t Price: %.2f, \t TransAmount  %.2f - %.2f %v \n", v.Advertiser.NickName, v.Adv.Price, v.Adv.MinSingleTransAmount, v.Adv.DynamicMaxSingleTransAmount, v.Adv.TradeMethods)
						msg += fmt.Sprintf("← Sell Adv: %s, Price: %.2f, Amount: %.2f - %.2f", v.Advertiser.NickName, v.Adv.Price, v.Adv.MinSingleTransAmount, v.Adv.DynamicMaxSingleTransAmount)
					}

					fmt.Println("++++++++++++++++++++++++++++++++++++++++++++")

					if barkId != "" {
						err = pkg.PushSuccess(msg, barkId)
						if err != nil {
							fmt.Println(err)
						}
					}
				}
			}
		}
	},
}
