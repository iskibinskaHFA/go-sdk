package migrate

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"time"
)

//Seed is seeding testing environment
func Seed(db *gorm.DB) {
	layout := "2006-01-02"
	periodStart, _ := time.Parse(layout, "2020-01-01")
	periodEnd, _ := time.Parse(layout, "2020-03-01")

	fmt.Println(periodStart)
	fmt.Println(periodEnd)

	db.Create(&Header{
		HeaderID:       []byte("123"),
		SenderHeaderID: "SenderHeaderID",
		SenderID:       "567",
		SenderName:     "Spotify",
		PeriodStart:    periodStart,
		PeriodEnd:      periodEnd,
	})

	rateFormulaCRB := getCRBFormula()
	rateFormulaDownload := getDownloadFormula()
	rateFormulaRingTone := getRingtoneFormula()

	createServices(db)

	// relations not working for postgres ?
	rf := RateFormula{
		RateFormulaID: 1,
		Formula:       postgres.Jsonb{RawMessage: rateFormulaCRB},
	}

	db.Save(&rf)

	db.Create(&RateFormula{
		RateFormulaID: 2,
		Formula:       postgres.Jsonb{RawMessage: rateFormulaDownload},
	})

	db.Create(&RateFormula{
		RateFormulaID: 3,
		Formula:       postgres.Jsonb{RawMessage: rateFormulaRingTone},
	})

	createRateDefinitions(db)
	createUsageSummary(db)
	createStepLogsDefinitions(db)
}

func getRingtoneFormula() json.RawMessage {
	return json.RawMessage(`
	{
	    "1.rate":  "{{.ringtone }}"
	}`)
}

func getDownloadFormula() json.RawMessage {
	return json.RawMessage(`
		{
	    "1.rate" :		 			"{{.fiveAndLess}}",
	    "2.rateGreaterThanFive" : 	"{{.moreThanFive}}"
		}
	`)
}

func getCRBFormula() json.RawMessage {
	return json.RawMessage(`
		{
			"1.recordLabelCost":  						   "Label_content_cost * {{.record}}",
	 		"2.subscriber": 				  	    		 "{{.subscriber}} * Subscriber_count",
			"3.lesserOfLabelCostAndSubscriber": 		 	 "min(recordLabelCost, maxIfZero(subscriber))",
			"4.musicServiceRevenue":						 "{{.rev}} * Net_service_revenue",
		    "5.allInOneRoyalty":		  					 "max(musicServiceRevenue, lesserOfLabelCostAndSubscriber)",
			"6.mechanicalRoyalty":					     "allInOneRoyalty - Performance_royalties",
			"7.adjustedSubscriberCount":						"{{.floor}} * Subscriber_count",
    	    "8.payableRoyalty":							 "max(mechanicalRoyalty, adjustedSubscriberCount)",
			"9.rate":									  	 "payableRoyalty/AdjustedUnitsTotal"
		}
	`)
}

func createUsageSummary(db *gorm.DB) {
	metadata := json.RawMessage(`
	{
	 "Net_service_revenue": 28062169.22,
	 "Subscriber_count": 23314946,
	 "Label_content_cost": 13722127.14,
	 "Performance_royalties": 1717547.37,
	 "AdjustedUnitsTotal":	8291.12

	}`)

	db.Create(&UsageSummary{
		UsageSummaryID: []byte("123"),
		ServiceID:      "1",
		HeaderID:       []byte("123"),
		SalesData:      postgres.Jsonb{RawMessage: metadata},
	})
}

func createRateDefinitions(db *gorm.DB) {
	layout := "2006-01-02"
	periodStart, _ := time.Parse(layout, "2020-01-01")
	periodEnd, _ := time.Parse(layout, "2020-03-01")

	rdfs := []RateDefinition{
		{RateDefinitionID: 1, ServiceID: "1", StartDate: periodStart, EndDate: periodEnd, RateFormulaID: 1,
			Constants: postgres.Jsonb{RawMessage: json.RawMessage(`{
					"record":  0.2410,
					"floor":   0.1500,
					"rev": 	   0.1330,
					"subscriber": 0 }`),
			},
		},
		{
			RateDefinitionID: 2, ServiceID: "2", StartDate: periodStart, EndDate: periodEnd, RateFormulaID: 1,
			Constants: postgres.Jsonb{RawMessage: json.RawMessage(`{
					"record":  0.2410,
					"floor":   0.3000,
					"rev": 	   0.1330,
					"subscriber": 0
  	 				}`),
			},
		},
		{
			RateDefinitionID: 3, ServiceID: "3", StartDate: periodStart, EndDate: periodEnd, RateFormulaID: 1,
			Constants: postgres.Jsonb{RawMessage: json.RawMessage(`{
				"record":  0.2410,
				"floor":   0.5000,
				"rev": 	   0.1330,
				"subscriber": 0
			}`),
			},
		},
		{
			RateDefinitionID: 7, ServiceID: "7", StartDate: periodStart, EndDate: periodEnd, RateFormulaID: 2,
			Constants: postgres.Jsonb{RawMessage: json.RawMessage(`{ "fiveAndLess"  : 0.0910, "moreThanFive" : 0.0875 }`)},
		},
		{
			RateDefinitionID: 8, ServiceID: "8", StartDate: periodStart, EndDate: periodEnd, RateFormulaID: 3,
			Constants: postgres.Jsonb{RawMessage: json.RawMessage(`{"ringtone" : 0.24 }`)},
		},
	}
	for _, rdf := range rdfs {
		db.Create(&rdf)
	}
}
func createServices(db *gorm.DB) {
	db.Create(&Service{
		ServiceID:   "1",
		Description: "CRB",
		Name:        "S1",
	})

	db.Create(&Service{
		ServiceID:   "2",
		Description: "CRB",
		Name:        "S3A",
	})

	db.Create(&Service{
		ServiceID:   "3",
		Description: "CRB",
		Name:        "S3",
	})

	db.Create(&Service{
		ServiceID:   "5",
		Description: "CRB",
		Name:        "S5",
	})

	db.Create(&Service{
		ServiceID:   "6",
		Description: "CRB",
		Name:        "S6",
	})

	db.Create(&Service{
		ServiceID:   "7",
		Description: "Download",
		Name:        "Download",
	})

	db.Create(&Service{
		ServiceID:   "8",
		Description: "Ringtone",
		Name:        "Ringtone",
	})

}

func createStepLogsDefinitions(db *gorm.DB) {
	logs := []CalcStepsLogDefinition{
		{
			ServiceID: "1", Result: "Subscriber_count", Sprintf: "Subscribers: %v", Step: "Inputs", SequenceID: 1,
		},
		{
			ServiceID: "1", Result: "Net_service_revenue", Sprintf: "Service Revenues: %v", Step: "Inputs", SequenceID: 2,
		},
		{
			ServiceID: "1", Result: "Label_content_cost", Sprintf: "Total Cost of Content: %v", Step: "Inputs", SequenceID: 3,
		},
		{
			ServiceID: "1", Result: "Performance_royalties", Sprintf: "Performance Royalties: %v", Step: "Inputs", SequenceID: 4,
		},
		{
			ServiceID: "1", Result: "AdjustedUnitsTotal", Sprintf: "Plays: %v", Step: "Inputs", SequenceID: 5,
		},
		{
			ServiceID: "1", Result: "musicServiceRevenue", Params: "rev", Sprintf: "a) %2.0f %% of Service Revenue", Step: "Step 1", SequenceID: 6,
		},
		{
			ServiceID: "1", Params: "record", Result: "recordLabelCost", Sprintf: "b) %2.0f %% of Total Cost of Content", Step: "Step 1", SequenceID: 7,
		},
		{
			ServiceID: "1", Result: "allInOneRoyalty", Sprintf: "All-In Royalty Pool equals : %v", Step: "Step 1", SequenceID: 8,
		},
		{
			ServiceID: "1", Result: "Performance_royalties", Sprintf: "Performance Royalties %v", Step: "Step 2", SequenceID: 9,
		},
		{
			ServiceID: "1", Params: "Performance_royalties", Result: "mechanicalRoyalty", Sprintf: "Separate Mechanical from Performance (less %v)",
			Step: "Step 2", SequenceID: 10,
		},
		{
			ServiceID: "1", Params: "floor", Result: "adjustedSubscriberCount", Sprintf: "$ %v/qualified subscriber/month", Step: "Step 2",
			SequenceID: 11,
		},
		{
			ServiceID: "1", Result: "mechanicalRoyalty", Sprintf: "Payable Mechanical Royalty Pool equals %v", Step: "Step 3", SequenceID: 12,
		},
		{
			ServiceID: "1", Result: "rate", Sprintf: "Effective Mechanical Rate per Play %v", Step: "Step 3", SequenceID: 13,
		},
	}
	for _, log := range logs {
		db.Create(&log)
	}
}
