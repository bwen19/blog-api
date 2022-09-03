package util

import (
	"fmt"
	"net/mail"
)

func ValidateString(value string, minLength int, maxLength int) error {
	length := len(value)
	if maxLength == 0 {
		if length < minLength {
			return fmt.Errorf("must contain at least %d characters", minLength)
		}
		return nil
	}
	if length < minLength || length > maxLength {
		return fmt.Errorf("must contain %d-%d characters", minLength, maxLength)
	}
	return nil
}

func ValidateEmail(value string) error {
	if err := ValidateString(value, 3, 50); err != nil {
		return err
	}
	if _, err := mail.ParseAddress(value); err != nil {
		return fmt.Errorf("invalid email address")
	}
	return nil
}

func ValidateNumber(value int64, minValue int64, maxValue int64) error {
	if maxValue == 0 {
		if value < minValue {
			return fmt.Errorf("must be at least %d", minValue)
		}
		return nil
	}
	if value < minValue || value > maxValue {
		return fmt.Errorf("must be between %d to %d", minValue, maxValue)
	}
	return nil
}

func ValidateID(ID int64) error {
	if ID < 1 {
		return fmt.Errorf("must be a positive integer")
	}
	return nil
}

func ValidateOneOf(value string, list []string) error {
	for _, v := range list {
		if value == v {
			return nil
		}
	}
	return fmt.Errorf("must be one of values: %s", list)
}

// -------------------------------------------------------------------

type Page interface {
	GetPageId() int32
	GetPageSize() int32
}

func ValidatePage(page Page) error {
	if err := ValidateNumber(int64(page.GetPageId()), 1, 0); err != nil {
		return fmt.Errorf("pageId: %s", err.Error())
	}
	if err := ValidateNumber(int64(page.GetPageSize()), 5, 50); err != nil {
		return fmt.Errorf("pageSize: %s", err.Error())
	}
	return nil
}

type Order interface {
	GetOrder() string
	GetOrderBy() string
}

func ValidateOrder(order Order, options []string) error {
	if err := ValidateOneOf(order.GetOrder(), []string{"asc", "desc"}); err != nil {
		return fmt.Errorf("order: %s", err.Error())
	}
	if err := ValidateOneOf(order.GetOrderBy(), options); err != nil {
		return fmt.Errorf("orderBy: %s", err.Error())
	}
	return nil
}

type PageOrder interface {
	Page
	Order
}

func ValidatePageOrder(pageOrder PageOrder, options []string) error {
	if err := ValidatePage(pageOrder); err != nil {
		return err
	}
	if err := ValidateOrder(pageOrder, options); err != nil {
		return err
	}
	return nil
}
