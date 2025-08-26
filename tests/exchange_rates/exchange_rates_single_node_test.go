package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

const (
	// Test equivalents from decimals map
	EQUIVALENT_101  = "101"  // 2 decimals
	EQUIVALENT_1001 = "1001" // 8 decimals
	EQUIVALENT_1002 = "1002" // 8 decimals
	EQUIVALENT_2002 = "2002" // 6 decimals

	// Test values for conversion validation
	TEST_VALUE_11207154      = "11207154"
	TEST_SHIFT_4             = int16(4)
	TEST_REAL_RATE_112071_54 = "112071.54"
	TEST_VALUE_112071        = "112071"
	TEST_SHIFT_2             = int16(2)
	TEST_REAL_RATE_112071    = "112071"
	TEST_VALUE_89228719      = "89228719"
	TEST_SHIFT_11            = int16(11)
	TEST_REAL_RATE_SMALL     = "0.0000089228719"

	// Edge case constants
	UNKNOWN_EQUIVALENT_1 = "999"         // Not in DecimalsMap
	UNKNOWN_EQUIVALENT_2 = "1111"        // Not in DecimalsMap
	MAX_INT16_SHIFT      = int16(32767)  // Maximum int16 value
	MIN_INT16_SHIFT      = int16(-32768) // Minimum int16 value

	// HTTP status codes
	HTTP_STATUS_OK             = 200
	HTTP_STATUS_BAD_REQUEST    = 400
	HTTP_STATUS_RATE_NOT_FOUND = 607 // Non-existent exchange rate

	// TTL test constants
	TTL_WAIT_DURATION = 5*time.Minute + 30*time.Second
)

var (
	exchangeRatesNextNodeIndex = 1
)

func getNextIPForExchangeRatesTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForOpenChannelTest, exchangeRatesNextNodeIndex)
	exchangeRatesNextNodeIndex++
	return ip
}

// Helper to create and run a single node for exchange rates tests
func setupNodeForExchangeRatesTest(t *testing.T) (*vtcp.Node, *vtcp.Cluster) {
	node := vtcp.NewNode(t, getNextIPForExchangeRatesTest(), "exchange-rates-node")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node})
	return node, cluster
}

// TestExchangeRatesValidationAndConversion covers test cases 1-12
func TestExchangeRatesValidationAndConversion(t *testing.T) {
	node, _ := setupNodeForExchangeRatesTest(t)

	// Test Case 1: Setting rate without specifying rate value → HTTP 400
	t.Run("Case1_MissingRateParameters", func(t *testing.T) {
		// Try to set rate without any rate parameters - this should cause HTTP 400
		node.SetExchangeRate(t, EQUIVALENT_1001, EQUIVALENT_2002, "", nil, nil, HTTP_STATUS_BAD_REQUEST)
		t.Logf("Case 1: Successfully validated missing rate parameters (HTTP 400)")
	})

	// Test Case 2: Setting rate with both real_rate and native (value+shift) → HTTP 400
	t.Run("Case2_ConflictingParameters", func(t *testing.T) {
		// Try to set rate with both real_rate and native parameters in single request - should cause HTTP 400
		node.SetExchangeRateWithConflictingParameters(t, EQUIVALENT_1001, EQUIVALENT_2002, TEST_REAL_RATE_112071_54, TEST_VALUE_11207154, TEST_SHIFT_4, nil, nil, HTTP_STATUS_BAD_REQUEST)
		t.Logf("Case 2: Successfully validated conflicting parameters in single request (HTTP 400)")
	})

	// Test Case 3: Setting real_rate with >16 decimal places → HTTP 400
	t.Run("Case3_ExcessiveDecimalPrecision", func(t *testing.T) {
		// Create a rate with more than 16 decimal places
		excessivePrecisionRate := "112071.12345678901234567"
		node.SetExchangeRate(t, EQUIVALENT_1001, EQUIVALENT_2002, excessivePrecisionRate, nil, nil, HTTP_STATUS_BAD_REQUEST)
		t.Logf("Case 3: Successfully validated excessive decimal precision (HTTP 400)")
	})

	// Test Case 4: Native input (1001→2002, value=11207154, shift=4) → real_rate=112071.54
	t.Run("Case4_NativeToRealHighPrecision", func(t *testing.T) {
		node.ClearExchangeRates(t)
		node.SetExchangeRateNative(t, EQUIVALENT_1001, EQUIVALENT_2002, TEST_VALUE_11207154, TEST_SHIFT_4, nil, nil, HTTP_STATUS_OK)
		rate := node.GetExchangeRate(t, EQUIVALENT_1001, EQUIVALENT_2002, HTTP_STATUS_OK)
		if rate.RealRate != TEST_REAL_RATE_112071_54 {
			t.Fatalf("Case 4: Expected real_rate=%s, got=%s", TEST_REAL_RATE_112071_54, rate.RealRate)
		}
		t.Logf("Case 4: Native to Real conversion successful: %s,%d → %s", TEST_VALUE_11207154, TEST_SHIFT_4, rate.RealRate)
	})

	// Test Case 5: Real input (1001→2002, real_rate=112071.54) → value=11207154, shift=4
	t.Run("Case5_RealToNativeHighPrecision", func(t *testing.T) {
		node.ClearExchangeRates(t)
		node.SetExchangeRate(t, EQUIVALENT_1001, EQUIVALENT_2002, TEST_REAL_RATE_112071_54, nil, nil, HTTP_STATUS_OK)
		rate := node.GetExchangeRate(t, EQUIVALENT_1001, EQUIVALENT_2002, HTTP_STATUS_OK)
		if rate.Value != TEST_VALUE_11207154 || rate.Shift != TEST_SHIFT_4 {
			t.Fatalf("Case 5: Expected value=%s, shift=%d, got value=%s, shift=%d",
				TEST_VALUE_11207154, TEST_SHIFT_4, rate.Value, rate.Shift)
		}
		t.Logf("Case 5: Real to Native conversion successful: %s → %s,%d", TEST_REAL_RATE_112071_54, rate.Value, rate.Shift)
	})

	// Test Case 6: Native input (1001→2002, value=112071, shift=2) → real_rate=112071
	t.Run("Case6_NativeToRealInteger", func(t *testing.T) {
		node.ClearExchangeRates(t)
		node.SetExchangeRateNative(t, EQUIVALENT_1001, EQUIVALENT_2002, TEST_VALUE_112071, TEST_SHIFT_2, nil, nil, HTTP_STATUS_OK)
		rate := node.GetExchangeRate(t, EQUIVALENT_1001, EQUIVALENT_2002, HTTP_STATUS_OK)
		if rate.RealRate != TEST_REAL_RATE_112071 {
			t.Fatalf("Case 6: Expected real_rate=%s, got=%s", TEST_REAL_RATE_112071, rate.RealRate)
		}
		t.Logf("Case 6: Native to Real integer conversion successful: %s,%d → %s", TEST_VALUE_112071, TEST_SHIFT_2, rate.RealRate)
	})

	// Test Case 7: Real input (1001→2002, real_rate=112071) → value=112071, shift=2
	t.Run("Case7_RealToNativeInteger", func(t *testing.T) {
		node.ClearExchangeRates(t)
		node.SetExchangeRate(t, EQUIVALENT_1001, EQUIVALENT_2002, TEST_REAL_RATE_112071, nil, nil, HTTP_STATUS_OK)
		rate := node.GetExchangeRate(t, EQUIVALENT_1001, EQUIVALENT_2002, HTTP_STATUS_OK)
		if rate.Value != TEST_VALUE_112071 || rate.Shift != TEST_SHIFT_2 {
			t.Fatalf("Case 7: Expected value=%s, shift=%d, got value=%s, shift=%d",
				TEST_VALUE_112071, TEST_SHIFT_2, rate.Value, rate.Shift)
		}
		t.Logf("Case 7: Real to Native integer conversion successful: %s → %s,%d", TEST_REAL_RATE_112071, rate.Value, rate.Shift)
	})

	// Test Case 8: Native input (2002→1001, value=89228719, shift=11) → real_rate=0.0000089228719
	t.Run("Case8_NativeToRealSmall", func(t *testing.T) {
		node.ClearExchangeRates(t)
		node.SetExchangeRateNative(t, EQUIVALENT_2002, EQUIVALENT_1001, TEST_VALUE_89228719, TEST_SHIFT_11, nil, nil, HTTP_STATUS_OK)
		rate := node.GetExchangeRate(t, EQUIVALENT_2002, EQUIVALENT_1001, HTTP_STATUS_OK)
		if rate.RealRate != TEST_REAL_RATE_SMALL {
			t.Fatalf("Case 8: Expected real_rate=%s, got=%s", TEST_REAL_RATE_SMALL, rate.RealRate)
		}
		t.Logf("Case 8: Native to Real small decimal conversion successful: %s,%d → %s", TEST_VALUE_89228719, TEST_SHIFT_11, rate.RealRate)
	})

	// Test Case 9: Real input (2002→1001, real_rate=0.0000089228719) → value=89228719, shift=11
	t.Run("Case9_RealToNativeSmall", func(t *testing.T) {
		node.ClearExchangeRates(t)
		node.SetExchangeRate(t, EQUIVALENT_2002, EQUIVALENT_1001, TEST_REAL_RATE_SMALL, nil, nil, HTTP_STATUS_OK)
		rate := node.GetExchangeRate(t, EQUIVALENT_2002, EQUIVALENT_1001, HTTP_STATUS_OK)
		if rate.Value != TEST_VALUE_89228719 || rate.Shift != TEST_SHIFT_11 {
			t.Fatalf("Case 9: Expected value=%s, shift=%d, got value=%s, shift=%d",
				TEST_VALUE_89228719, TEST_SHIFT_11, rate.Value, rate.Shift)
		}
		t.Logf("Case 9: Real to Native small decimal conversion successful: %s → %s,%d", TEST_REAL_RATE_SMALL, rate.Value, rate.Shift)
	})

	// Test Case 10: Getting non-existent exchange rate → HTTP 607
	t.Run("Case10_NonExistentRateQuery", func(t *testing.T) {
		node.ClearExchangeRates(t)
		// Try to get a rate that doesn't exist - expect HTTP 607
		rate := node.GetExchangeRate(t, EQUIVALENT_1001, EQUIVALENT_2002, HTTP_STATUS_RATE_NOT_FOUND)
		if rate != nil {
			t.Fatalf("Case 10: Expected nil for non-existent rate, but got: %+v", rate)
		}
		t.Logf("Case 10: Non-existent rate query correctly returned HTTP 607")
	})

	// Test Case 11: Setting rate with maximum int16 shift value (32767) → correct conversion
	t.Run("Case11_MaximumShiftBoundary", func(t *testing.T) {
		node.ClearExchangeRates(t)
		node.SetExchangeRateNative(t, EQUIVALENT_1001, EQUIVALENT_2002, "123", MAX_INT16_SHIFT, nil, nil, HTTP_STATUS_OK)
		rate := node.GetExchangeRate(t, EQUIVALENT_1001, EQUIVALENT_2002, HTTP_STATUS_OK)
		if rate.Shift != MAX_INT16_SHIFT {
			t.Fatalf("Case 11: Expected shift=%d, got shift=%d", MAX_INT16_SHIFT, rate.Shift)
		}
		t.Logf("Case 11: Maximum shift boundary test successful: shift=%d", rate.Shift)
	})

	// Test Case 12: Setting rate with minimum int16 shift value (-32768) → correct conversion
	t.Run("Case12_MinimumShiftBoundary", func(t *testing.T) {
		node.ClearExchangeRates(t)
		node.SetExchangeRateNative(t, EQUIVALENT_1002, EQUIVALENT_2002, "456", MIN_INT16_SHIFT, nil, nil, HTTP_STATUS_OK)
		rate := node.GetExchangeRate(t, EQUIVALENT_1002, EQUIVALENT_2002, HTTP_STATUS_OK)
		if rate.Shift != MIN_INT16_SHIFT {
			t.Fatalf("Case 12: Expected shift=%d, got shift=%d", MIN_INT16_SHIFT, rate.Shift)
		}
		t.Logf("Case 12: Minimum shift boundary test successful: shift=%d", rate.Shift)
	})

	// Clean up after all tests
	node.ClearExchangeRates(t)
}

// TestExchangeRatesCRUD covers test case 13
func TestExchangeRatesCRUD(t *testing.T) {
	node, _ := setupNodeForExchangeRatesTest(t)

	// Set multiple rates for different equivalent pairs
	t.Run("SetMultipleRates", func(t *testing.T) {
		node.SetExchangeRate(t, EQUIVALENT_101, EQUIVALENT_1001, "100.50", nil, nil, HTTP_STATUS_OK)
		node.SetExchangeRate(t, EQUIVALENT_1001, EQUIVALENT_1002, "1.25", nil, nil, HTTP_STATUS_OK)
		node.SetExchangeRate(t, EQUIVALENT_1002, EQUIVALENT_2002, "0.75", nil, nil, HTTP_STATUS_OK)
		t.Logf("Set 3 different exchange rates successfully")
	})

	// List all rates and verify count
	t.Run("ListAndVerifyCount", func(t *testing.T) {
		rates := node.ListExchangeRates(t)
		if rates.Count != 3 {
			t.Fatalf("Expected 3 rates, got %d", rates.Count)
		}
		if len(rates.Rates) != 3 {
			t.Fatalf("Expected 3 rates in list, got %d", len(rates.Rates))
		}
		t.Logf("Listed all rates successfully: count=%d", rates.Count)
	})

	// Delete one specific rate
	t.Run("DeleteSpecificRate", func(t *testing.T) {
		node.DeleteExchangeRate(t, EQUIVALENT_1001, EQUIVALENT_1002)
		rates := node.ListExchangeRates(t)
		if rates.Count != 2 {
			t.Fatalf("Expected 2 rates after deletion, got %d", rates.Count)
		}

		// Verify the deleted rate is not in the list
		for _, rate := range rates.Rates {
			if rate.EquivalentFrom == EQUIVALENT_1001 && rate.EquivalentTo == EQUIVALENT_1002 {
				t.Fatalf("Deleted rate still found in list")
			}
		}
		t.Logf("Successfully deleted specific rate, remaining count: %d", rates.Count)
	})

	// Verify other rates remain
	t.Run("VerifyOtherRatesRemain", func(t *testing.T) {
		rates := node.ListExchangeRates(t)
		foundRate1 := false
		foundRate3 := false

		for _, rate := range rates.Rates {
			if rate.EquivalentFrom == EQUIVALENT_101 && rate.EquivalentTo == EQUIVALENT_1001 {
				foundRate1 = true
			}
			if rate.EquivalentFrom == EQUIVALENT_1002 && rate.EquivalentTo == EQUIVALENT_2002 {
				foundRate3 = true
			}
		}

		if !foundRate1 || !foundRate3 {
			t.Fatalf("Expected remaining rates not found. foundRate1=%t, foundRate3=%t", foundRate1, foundRate3)
		}
		t.Logf("Verified other rates remain after selective deletion")
	})

	// Clear all rates
	t.Run("ClearAllRates", func(t *testing.T) {
		node.ClearExchangeRates(t)
		rates := node.ListExchangeRates(t)
		if rates.Count != 0 {
			t.Fatalf("Expected 0 rates after clearing, got %d", rates.Count)
		}
		if len(rates.Rates) != 0 {
			t.Fatalf("Expected 0 rates in list after clearing, got %d", len(rates.Rates))
		}
		t.Logf("Successfully cleared all rates, count: %d", rates.Count)
	})
}

// TestExchangeRatesTTL covers test case 14
func TestExchangeRatesTTL(t *testing.T) {
	node, _ := setupNodeForExchangeRatesTest(t)

	// Set one exchange rate
	t.Run("SetRateForTTL", func(t *testing.T) {
		node.SetExchangeRate(t, EQUIVALENT_1001, EQUIVALENT_2002, TEST_REAL_RATE_112071_54, nil, nil, HTTP_STATUS_OK)
		t.Logf("Set exchange rate for TTL test")
	})

	// Verify rate is accessible immediately
	t.Run("VerifyImmediateAccess", func(t *testing.T) {
		rate := node.GetExchangeRate(t, EQUIVALENT_1001, EQUIVALENT_2002, HTTP_STATUS_OK)
		if rate.RealRate != TEST_REAL_RATE_112071_54 {
			t.Fatalf("Rate not accessible immediately after setting. Expected %s, got %s",
				TEST_REAL_RATE_112071_54, rate.RealRate)
		}
		t.Logf("Rate is accessible immediately after setting")
	})

	// Wait for TTL expiration
	t.Run("WaitForTTLExpiration", func(t *testing.T) {
		t.Logf("Waiting for TTL expiration: %v (approximately 5.5 minutes)", TTL_WAIT_DURATION)
		startTime := time.Now()
		time.Sleep(TTL_WAIT_DURATION)
		elapsed := time.Since(startTime)
		t.Logf("Waited for %v", elapsed)
	})

	// Verify rate has expired
	t.Run("VerifyRateExpired", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				// Rate should be expired, so getting it should fail
				t.Logf("Rate correctly expired after TTL - GetExchangeRate failed as expected: %v", r)
				return
			}
			// If we get here without panic, the rate might still exist, which would be a test failure
			rate := node.GetExchangeRate(t, EQUIVALENT_1001, EQUIVALENT_2002, HTTP_STATUS_RATE_NOT_FOUND)
			if rate != nil {
				t.Fatalf("Rate should have expired after TTL wait but is still accessible: %+v", rate)
			}
		}()

		// Try to get the rate - this should fail due to TTL expiration
		node.GetExchangeRate(t, EQUIVALENT_1001, EQUIVALENT_2002, HTTP_STATUS_RATE_NOT_FOUND)
	})
}
