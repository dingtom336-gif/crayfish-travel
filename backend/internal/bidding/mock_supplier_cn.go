package bidding

import (
	"fmt"
	"math/rand"
)

// MockSupplierCN provides Chinese mock data as fallback when FlyAI is unavailable.
type MockSupplierCN struct{}

// NewMockSupplierCN creates a Chinese mock supplier.
func NewMockSupplierCN() *MockSupplierCN {
	return &MockSupplierCN{}
}

var mockPackagesCN = []struct {
	titleTemplate string
	hotel         string
	highlights    []string
	inclusions    []string
	starRating    float64
	imageURL      string
}{
	{
		titleTemplate: "%s%d天%d晚 海滨度假天堂",
		hotel:         "希尔顿度假酒店",
		highlights:    []string{"一线海景房", "专车接送机", "私人管家服务"},
		inclusions:    []string{"含早餐", "SPA 体验券", "儿童乐园"},
		starRating:    4.8,
		imageURL:      "https://img.alicdn.com/imgextra/i1/6000000007629/O1CN01example01.jpg",
	},
	{
		titleTemplate: "%s%d天%d晚 亲子欢乐游",
		hotel:         "万豪亲子度假酒店",
		highlights:    []string{"水上乐园免费玩", "家庭套房", "专属导游"},
		inclusions:    []string{"含三餐", "主题乐园门票", "亲子摄影"},
		starRating:    4.6,
		imageURL:      "https://img.alicdn.com/imgextra/i2/6000000007629/O1CN01example02.jpg",
	},
	{
		titleTemplate: "%s%d天%d晚 奢华逸享",
		hotel:         "丽思卡尔顿",
		highlights:    []string{"总统套房", "私人泳池", "米其林餐厅"},
		inclusions:    []string{"全包服务", "豪车接送", "私人厨师"},
		starRating:    4.9,
		imageURL:      "https://img.alicdn.com/imgextra/i3/6000000007629/O1CN01example03.jpg",
	},
	{
		titleTemplate: "%s%d天%d晚 经济优选",
		hotel:         "全季酒店",
		highlights:    []string{"市中心位置", "免费 WiFi", "穿梭巴士"},
		inclusions:    []string{"含早餐", "城市地图", "欢迎饮品"},
		starRating:    4.2,
		imageURL:      "https://img.alicdn.com/imgextra/i4/6000000007629/O1CN01example04.jpg",
	},
	{
		titleTemplate: "%s%d天%d晚 文化深度游",
		hotel:         "凯悦酒店",
		highlights:    []string{"文化遗产探访", "地道美食体验", "古迹参观"},
		inclusions:    []string{"含半餐", "专业导览", "纪念品礼包"},
		starRating:    4.5,
		imageURL:      "https://img.alicdn.com/imgextra/i5/6000000007629/O1CN01example05.jpg",
	},
	{
		titleTemplate: "%s%d天%d晚 探险之旅",
		hotel:         "喜来登探险山庄",
		highlights:    []string{"浮潜体验", "徒步路线", "日落巡航"},
		inclusions:    []string{"装备租赁", "午餐包", "退改保障"},
		starRating:    4.4,
		imageURL:      "https://img.alicdn.com/imgextra/i6/6000000007629/O1CN01example06.jpg",
	},
	{
		titleTemplate: "%s%d天%d晚 浪漫之旅",
		hotel:         "四季酒店",
		highlights:    []string{"情侣 SPA", "烛光晚餐", "海滨别墅"},
		inclusions:    []string{"全餐服务", "香槟礼遇", "专属摄影"},
		starRating:    4.7,
		imageURL:      "https://img.alicdn.com/imgextra/i7/6000000007629/O1CN01example07.jpg",
	},
	{
		titleTemplate: "%s%d天%d晚 超值套餐",
		hotel:         "万怡酒店",
		highlights:    []string{"泳池开放", "健身中心", "商务中心"},
		inclusions:    []string{"含早餐", "延迟退房", "免费停车"},
		starRating:    4.3,
		imageURL:      "https://img.alicdn.com/imgextra/i8/6000000007629/O1CN01example08.jpg",
	},
}

// FetchQuotes returns Chinese mock quotes.
func (m *MockSupplierCN) FetchQuotes(destination string, days int, budgetCents int64, _, _ int) ([]Quote, error) {
	nights := days - 1
	if nights < 1 {
		nights = 1
	}

	const refundGuaranteeFee int64 = 10000

	quotes := make([]Quote, len(mockPackagesCN))
	for i, pkg := range mockPackagesCN {
		priceMultiplier := 0.8 + rand.Float64()*0.4
		basePrice := int64(float64(budgetCents) * priceMultiplier)
		totalPrice := basePrice + refundGuaranteeFee

		quotes[i] = Quote{
			Supplier:                "飞猪",
			PackageTitle:            fmt.Sprintf(pkg.titleTemplate, destination, days, nights),
			Destination:             destination,
			DurationDays:            days,
			DurationNights:          nights,
			BasePriceCents:          basePrice,
			RefundGuaranteeFeeCents: refundGuaranteeFee,
			TotalPriceCents:         totalPrice,
			StarRating:              pkg.starRating,
			ReviewCount:             50 + rand.Intn(200),
			HotelName:               pkg.hotel,
			Highlights:              pkg.highlights,
			Inclusions:              pkg.inclusions,
			ImageURL:                pkg.imageURL,
		}
	}

	return quotes, nil
}
