package account

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/oldfritter/lucy/dom"
)

const (
	FunCredit = 1  // 资金增加
	FunDebit  = -1 // 资金减少
)

var (
	ErrInsufficientBalance = errors.New("余额不足")
	ErrInvalidAmount       = errors.New("金额必须大于 0")
)

// AddBalance 为指定账户增加资金，记录资金变动来源
//
//	tx — 外部传入的事务（调用方负责 Commit / Rollback）
//	accountId — 目标账户 ID
//	amount — 增加金额（正数）
//	modifiableType — 变动来源类型（如 "Order"、"Refund"）
//	modifiableId — 变动来源记录 ID
//
//	在同一 (modifiableType, modifiableId, accountId) 组合上重复调用会返回错误（唯一约束），
//	防止同一笔来源被重复入账。
func AddBalance(tx *gorm.DB, accountId int, amount int, modifiableType string, modifiableId int) error {
	return changeBalance(tx, accountId, amount, FunCredit, modifiableType, modifiableId)
}

// DeductBalance 从指定账户扣除资金，记录资金变动来源
//
//	当余额不足时返回 ErrInsufficientBalance。
//	其余行为同 AddBalance。
func DeductBalance(tx *gorm.DB, accountId int, amount int, modifiableType string, modifiableId int) error {
	return changeBalance(tx, accountId, amount, FunDebit, modifiableType, modifiableId)
}

// changeBalance 通用资金变更：锁行读取 → 校验 → 写入版本记录 → 更新余额
func changeBalance(tx *gorm.DB, accountId int, amount int, fun int, modifiableType string, modifiableId int) error {
	if amount <= 0 {
		return fmt.Errorf("%w: %d", ErrInvalidAmount, amount)
	}

	// 1. 行级锁读取账户（SELECT … FOR UPDATE），防止并发修改
	var account dom.Account
	if err := tx.Raw("SELECT * FROM account WHERE id = ? FOR UPDATE", accountId).
		Scan(&account).Error; err != nil {
		return fmt.Errorf("锁定账户失败: %w", err)
	}
	if account.Id == 0 {
		return fmt.Errorf("账户 %d 不存在", accountId)
	}

	// 2. 校验余额（仅扣款时）
	delta := fun * amount
	if account.Available+delta < 0 {
		return ErrInsufficientBalance
	}

	// 3. 创建资金变动版本记录
	version := dom.AccountVersion{
		ModifiableType: modifiableType,
		ModifiableId:   modifiableId,
		AccountId:      accountId,
		Fun:            fun,
		Available:      amount, // 绝对值
	}
	if err := tx.Create(&version).Error; err != nil {
		return fmt.Errorf("记录资金变动失败: %w", err)
	}

	// 4. 更新账户余额
	newAvailable := account.Available + delta
	if err := tx.Model(&dom.Account{}).Where("id = ?", accountId).
		Update("available", newAvailable).Error; err != nil {
		return fmt.Errorf("更新余额失败: %w", err)
	}

	return nil
}
