package engine

import (
	"context"
	"encoding/json"

	"github.com/copernet/whc.go/btcjson"
	"github.com/copernet/whccommon/log"
	common "github.com/copernet/whccommon/model"
	"github.com/copernet/whcengine/model"
	"github.com/copernet/whcengine/util"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

var handlers map[uint64]TransactionHandler

func init() {
	handlers = map[uint64]TransactionHandler{
		0:   (*SimpleSendHandller)(nil),
		1:   (*BuyCrowdTokenHandller)(nil),
		3:   (*SendToOwnersHandler)(nil),
		4:   (*SendAllHandler)(nil),
		50:  (*CreateFixedTokenHandler)(nil),
		51:  (*CreateCrowdTokenHandler)(nil),
		53:  (*CloseCrowdTokenHandler)(nil),
		54:  (*CreateMangeTokenHandler)(nil),
		55:  (*GrantTokenHandler)(nil),
		56:  (*RevokeTokenHandler)(nil),
		68:  (*BurnBCHHandller)(nil),
		70:  (*ChangeTokenIssuerHandler)(nil),
		185: (*FrozenMangeTokenHandler)(nil),
		186: (*FrozenMangeTokenHandler)(nil),
	}

}

func BuildFactory(txtype uint64) TransactionHandler {
	return handlers[txtype]
}

type TransactionHandler interface {
	Invoke(t *btcjson.GenerateTransactionResult, txId int, amount *decimal.Decimal, ctx context.Context) (*common.AddressesInTx, error)
}

type BuyCrowdTokenHandller struct{}

func (h *BuyCrowdTokenHandller) Invoke(t *btcjson.GenerateTransactionResult, txId int, amount *decimal.Decimal, ctx context.Context) (*common.AddressesInTx, error) {
	//logic for whc
	neg := amount.Neg()
	txf := &common.AddressesInTx{Address: t.SendingAddress, PropertyID: t.PropertyID, Protocol: common.Wormhole, TxID: txId, AddressRole: common.Sender, BalanceAvailableCreditDebit: &neg, AddressTxIndex: 1}
	txt := &common.AddressesInTx{Address: t.ReferenceAddress, PropertyID: t.PropertyID, Protocol: common.Wormhole, TxID: txId, AddressRole: common.Recipient, BalanceAvailableCreditDebit: amount, AddressTxIndex: 1}
	model.InsertAddressInTx(txf, ctx)
	model.InsertAddressInTx(txt, ctx)

	tx := &common.AddressesInTx{Address: t.SendingAddress, PropertyID: t.PropertyID, Protocol: common.Wormhole, TxID: txId, AddressRole: common.Participant, BalanceAvailableCreditDebit: amount, AddressTxIndex: 1}
	if !t.Valid {
		return tx, nil
	}

	//update whc balance
	updateBalance(txf, ctx)
	updateBalance(txt, ctx)

	//update the crowdsale property tokens
	val := getAmount(t.PurchasedTokens)
	tx.BalanceAvailableCreditDebit = val
	tx.PropertyID = t.PurchasedPropertyID
	t.PropertyID = t.PurchasedPropertyID
	return tx, upsertProperty(t, ctx, false)
}

type ChangeTokenIssuerHandler struct{}

func (h *ChangeTokenIssuerHandler) Invoke(t *btcjson.GenerateTransactionResult, txId int, amount *decimal.Decimal, ctx context.Context) (*common.AddressesInTx, error) {
	tx := &common.AddressesInTx{Address: t.SendingAddress, PropertyID: t.PropertyID, Protocol: common.Wormhole, TxID: txId, AddressRole: common.Sender, BalanceAvailableCreditDebit: amount, AddressTxIndex: 1}
	if t.Valid {
		txs := &common.AddressesInTx{Address: t.SendingAddress, PropertyID: t.PropertyID, Protocol: common.Wormhole, TxID: txId, AddressRole: common.Issuer, BalanceAvailableCreditDebit: &decimal.Zero, AddressTxIndex: 1}
		model.InsertAddressInTx(txs, ctx)
		upsertProperty(t, ctx, false)

		tx.Address = t.ReferenceAddress
		tx.AddressRole = common.Recipient
	}

	return tx, nil
}

type RevokeTokenHandler struct{}

func (h *RevokeTokenHandler) Invoke(t *btcjson.GenerateTransactionResult, txId int, amount *decimal.Decimal, ctx context.Context) (*common.AddressesInTx, error) {
	neg := amount.Neg()
	tx := &common.AddressesInTx{Address: t.SendingAddress, PropertyID: t.PropertyID, Protocol: common.Wormhole, TxID: txId, AddressRole: common.Issuer, BalanceAvailableCreditDebit: &neg, AddressTxIndex: 1}
	return tx, upsertProperty(t, ctx, false)
}

type GrantTokenHandler struct{}

func (h *GrantTokenHandler) Invoke(t *btcjson.GenerateTransactionResult, txId int, amount *decimal.Decimal, ctx context.Context) (*common.AddressesInTx, error) {
	tx := &common.AddressesInTx{Address: t.SendingAddress, PropertyID: t.PropertyID, Protocol: common.Wormhole, TxID: txId, AddressRole: common.Issuer, BalanceAvailableCreditDebit: amount, AddressTxIndex: 1}
	if t.ReferenceAddress != "" && t.ReferenceAddress != t.SendingAddress {
		txs := &common.AddressesInTx{Address: t.ReferenceAddress, PropertyID: t.PropertyID, Protocol: common.Wormhole, TxID: txId, AddressRole: common.Recipient, BalanceAvailableCreditDebit: amount, AddressTxIndex: 1}
		model.InsertAddressInTx(txs, ctx)
		if t.Valid {
			updateBalance(txs, ctx)
		}
		tx.BalanceAvailableCreditDebit = &decimal.Zero
	}

	return tx, upsertProperty(t, ctx, false)
}

type CreateMangeTokenHandler struct{}

func (h *CreateMangeTokenHandler) Invoke(t *btcjson.GenerateTransactionResult, txId int, amount *decimal.Decimal, ctx context.Context) (*common.AddressesInTx, error) {
	//Create Manage,reset amount=0
	return createTokenOperate(t, txId, &decimal.Zero, ctx, true)
}

type CreateCrowdTokenHandler struct{}

func (h *CreateCrowdTokenHandler) Invoke(t *btcjson.GenerateTransactionResult, txId int, amount *decimal.Decimal, ctx context.Context) (*common.AddressesInTx, error) {
	//Create Crowd,reset amount=0,for This Tx,amount only limit Participation CorwdSale,not to owners
	return createTokenOperate(t, txId, &decimal.Zero, ctx, true)
}

type CloseCrowdTokenHandler struct{}

func (h *CloseCrowdTokenHandler) Invoke(t *btcjson.GenerateTransactionResult, txId int, amount *decimal.Decimal, ctx context.Context) (*common.AddressesInTx, error) {
	if t.Valid {
		log.WithCtx(ctx).Infof("corwdSale pid:%d has closed", t.PropertyID)

		//Due to crowdale deadline time or Manual closing crowdale,retrieve tokens to issuer.
		upsertProperty(t, ctx, false)
		retrieveTokens(t.PropertyID, txId, ctx)
	}

	return &common.AddressesInTx{Address: t.SendingAddress, PropertyID: t.PropertyID, Protocol: common.Wormhole, TxID: txId, AddressRole: common.Issuer, BalanceAvailableCreditDebit: &decimal.Zero, AddressTxIndex: 1}, nil
}

func retrieveTokens(pid int64, txId int, ctx context.Context) error {
	sp, err := model.GetSmartProperty(pid, common.Wormhole, ctx)
	if gorm.IsRecordNotFoundError(err) {
		return errors.New("can't find Smart_Property")
	}

	log.WithCtx(ctx).Infof("retrieve tokens,pid:%d,data:%s", sp.PropertyID, sp.PropertyData)

	var data map[string]interface{}
	err = json.Unmarshal([]byte(sp.PropertyData), &data)
	if err != nil {
		return err
	}

	//retrieve addedissuertokens
	val := data["addedissuertokens"]
	if val == nil || val.(string) == "" {
		return nil
	}
	fixVal, _ := decimal.NewFromString(val.(string))
	tx := &common.AddressesInTx{Address: sp.Issuer, PropertyID: sp.PropertyID, Protocol: common.Wormhole, TxID: txId, AddressRole: common.Payee, BalanceAvailableCreditDebit: &fixVal, AddressTxIndex: 1}
	if txId > 0 {
		model.InsertAddressInTx(tx, ctx)
	}

	updateBalance(tx, ctx)

	return nil
}

type CreateFixedTokenHandler struct{}

func (h *CreateFixedTokenHandler) Invoke(t *btcjson.GenerateTransactionResult, txId int, amount *decimal.Decimal, ctx context.Context) (*common.AddressesInTx, error) {
	return createTokenOperate(t, txId, amount, ctx, true)
}
func createTokenOperate(t *btcjson.GenerateTransactionResult, txId int, amount *decimal.Decimal, ctx context.Context, create bool) (*common.AddressesInTx, error) {
	fee := decimal.NewFromFloat(1)
	neg := fee.Neg()
	tx := &common.AddressesInTx{Address: t.SendingAddress, PropertyID: 1, Protocol: common.Wormhole, TxID: txId, AddressRole: common.Feepayer, BalanceAvailableCreditDebit: &neg, AddressTxIndex: 1}
	//model.InsertAddressInTx(tx, ctx)
	if t.Valid {
		updateBalance(tx, ctx)
	}
	t.TotalStoFee = fee.String()

	return &common.AddressesInTx{Address: t.SendingAddress, PropertyID: t.PropertyID, Protocol: common.Wormhole, TxID: txId, AddressRole: common.Issuer, BalanceAvailableCreditDebit: amount, AddressTxIndex: 1}, upsertProperty(t, ctx, create)

}

type SendAllHandler struct{}

func (h *SendAllHandler) Invoke(t *btcjson.GenerateTransactionResult, txId int, amount *decimal.Decimal, ctx context.Context) (*common.AddressesInTx, error) {

	for _, send := range t.SubSends {
		value := getAmount(send.Amount)

		//debit the sender
		neg := value.Neg()
		txs := &common.AddressesInTx{Address: t.SendingAddress, PropertyID: send.PropertyID, Protocol: common.Wormhole, TxID: txId, AddressRole: common.Sender, BalanceAvailableCreditDebit: &neg}
		model.InsertAddressInTx(txs, ctx)

		//credit the receiver
		txr := &common.AddressesInTx{Address: t.ReferenceAddress, PropertyID: send.PropertyID, Protocol: common.Wormhole, TxID: txId, AddressRole: common.Recipient, BalanceAvailableCreditDebit: value}
		model.InsertAddressInTx(txr, ctx)

		if t.Valid {
			updateBalance(txs, ctx)
			updateBalance(txr, ctx)
		}

		insertPropertyHistory(send.PropertyID, txId, ctx)
	}

	return nil, nil
}

type SendToOwnersHandler struct{}

func (h *SendToOwnersHandler) Invoke(t *btcjson.GenerateTransactionResult, txId int, amount *decimal.Decimal, ctx context.Context) (*common.AddressesInTx, error) {

	if t.Valid {
		filter := "*"
		sto, err := client.WhcGetSto(t.TxID, &filter)
		if err != nil {
			return nil, err
		}

		stoFee := getAmount(sto.TotalStoFee)
		neg := stoFee.Neg()
		tx := &common.AddressesInTx{Address: t.SendingAddress, PropertyID: 1, Protocol: common.Wormhole, TxID: txId, AddressRole: common.Feepayer, BalanceAvailableCreditDebit: &neg, AddressTxIndex: 1}
		model.InsertAddressInTx(tx, ctx)

		t.TotalStoFee = stoFee.String()
		t.Recipients = sto.Recipients

		updateBalance(tx, ctx)
		updateRecipients(sto, txId, ctx)
	}

	insertPropertyHistory(t.PropertyID, txId, ctx)

	neg := amount.Neg()
	return &common.AddressesInTx{AddressRole: common.Payer, BalanceAvailableCreditDebit: &neg, Address: t.SendingAddress, PropertyID: t.PropertyID, Protocol: common.Wormhole, TxID: txId, AddressTxIndex: 0}, nil
}
func updateRecipients(t *btcjson.GenerateTransactionResult, txId int, ctx context.Context) {
	var txindex int16 = 0
	for _, reci := range t.Recipients {
		balance := getAmount(reci.Amount)
		tx := &common.AddressesInTx{Address: reci.Address, PropertyID: t.PropertyID, Protocol: common.Wormhole, TxID: txId, AddressRole: common.Payee, BalanceAvailableCreditDebit: balance, AddressTxIndex: txindex}
		model.InsertAddressInTx(tx, ctx)

		updateBalance(tx, ctx)
		txindex += 1
	}
}

type SimpleSendHandller struct{}

func (h *SimpleSendHandller) Invoke(t *btcjson.GenerateTransactionResult, txId int, amount *decimal.Decimal, ctx context.Context) (*common.AddressesInTx, error) {
	neg := amount.Neg()
	tx := &common.AddressesInTx{Address: t.SendingAddress, PropertyID: t.PropertyID, Protocol: common.Wormhole, TxID: txId, AddressRole: common.Sender, BalanceAvailableCreditDebit: &neg}
	model.InsertAddressInTx(tx, ctx)

	if t.Valid {
		updateBalance(tx, ctx)
	}

	recvTx := &common.AddressesInTx{
		AddressRole:                 common.Recipient,
		BalanceAvailableCreditDebit: amount,
		Address:                     t.SendingAddress,
		PropertyID:                  t.PropertyID,
		Protocol:                    common.Wormhole,
		TxID:                        txId,
		AddressTxIndex:              0,
	}

	insertPropertyHistory(t.PropertyID, txId, ctx)

	if t.ReferenceAddress != "" {
		recvTx.Address = t.ReferenceAddress
		return recvTx, nil
	}

	// ReferenceAddress is empty,record account to Sender
	return recvTx, nil
}

func insertPropertyHistory(pid int64, txId int, ctx context.Context) {
	ph := &common.PropertyHistory{PropertyID: pid, TxID: int64(txId)}
	model.InsertPropertyHistory(ph, ctx)
}

func updateBalance(tx *common.AddressesInTx, ctx context.Context) {
	balance := model.GetAddressBalance(tx.Address, tx.Protocol, tx.PropertyID, ctx)
	vo := common.BalanceNotify{Address: tx.Address, PropertyID: tx.PropertyID, TxID: tx.TxID}
	bys, _ := json.Marshal(vo)
	if balance == nil {
		balance = &common.AddressBalance{
			Address:          tx.Address,
			PropertyID:       tx.PropertyID,
			Protocol:         tx.Protocol,
			LastTxID:         tx.TxID,
			Ecosystem:        common.Production,
			BalanceFrozen:    tx.BalanceFrozenCreditDebit,
			BalanceAvailable: tx.BalanceAvailableCreditDebit,
			BalanceReserved:  tx.BalanceReservedCreditDebit,
			BalanceAccepted:  tx.BalanceAcceptedCreditDebit,
		}

		model.InsertAddressBalance(balance, ctx)
		model.PushStack(model.AddressBalanceTip, string(bys), ctx)
		return
	}

	//Rest balance data
	if tx.BalanceAvailableCreditDebit != nil {
		*balance.BalanceAvailable = balance.BalanceAvailable.Add(*tx.BalanceAvailableCreditDebit)
	}

	if tx.BalanceReservedCreditDebit != nil {
		*balance.BalanceReserved = balance.BalanceReserved.Add(*tx.BalanceReservedCreditDebit)
	}

	if tx.BalanceAcceptedCreditDebit != nil {
		*balance.BalanceAccepted = balance.BalanceAccepted.Add(*tx.BalanceAcceptedCreditDebit)
	}

	if tx.BalanceFrozenCreditDebit != nil {
		*balance.BalanceFrozen = balance.BalanceFrozen.Add(*tx.BalanceFrozenCreditDebit)
	}

	balance.LastTxID = tx.TxID
	model.UpdateAddressBalance(balance, ctx)

	info, err := client.GetInfo()
	if err != nil {
		log.WithCtx(ctx).Errorf("GetInfo error:%s", err.Error())
		return
	}

	block := model.GetLastBlock(ctx)

	if int64(info.Blocks) > block.BlockHeight {
		log.WithCtx(ctx).Infof("Sync Not Finish.core height:%d,db height:%d", int64(info.Blocks), block.BlockHeight)
		return
	}

	model.PushStack(model.AddressBalanceTip, string(bys), ctx)
}

type BurnBCHHandller struct{}

func (h *BurnBCHHandller) Invoke(t *btcjson.GenerateTransactionResult, txId int, amount *decimal.Decimal, ctx context.Context) (*common.AddressesInTx, error) {
	tx := &common.AddressesInTx{Address: t.SendingAddress, PropertyID: t.PropertyID, Protocol: common.Wormhole,
		TxID: txId, AddressRole: common.Buyer, BalanceFrozenCreditDebit: amount, AddressTxIndex: 1}
	return tx, upsertProperty(t, ctx, false)
}
func upsertProperty(t *btcjson.GenerateTransactionResult, ctx context.Context, create bool) error {
	sp, ph, err := FetchPropertyAndHistory(t, ctx, create)
	if err != nil {
		return err
	}

	if sp != nil && ph != nil {
		model.UpsertSmartProperty(sp, ctx)
		model.InsertPropertyHistory(ph, ctx)
	}

	return nil
}

func FetchPropertyAndHistory(t *btcjson.GenerateTransactionResult, ctx context.Context, create bool) (*common.SmartProperty, *common.PropertyHistory, error) {
	if !t.Valid || t.PropertyID == 0 {
		return nil, nil, nil
	}

	pid := uint64(t.PropertyID)
	p, err := client.WhcGetProperty(pid)
	if err != nil {
		return nil, nil, err
	}

	lastTxId := GetTxByTxHash(t.TxID, ctx)
	createTxId := GetTxByTxHash(p.CreateTxID, ctx)

	base := util.To_Map(p)
	filter := true
	crowd, err := client.WhcGetCrowdSale(pid, &filter)
	if err == nil {
		err := patchForCrowd(crowd, base)
		if err != nil {
			return nil, nil, err
		}

		if create {
			crowd.Active = true
		}

		base = util.Merge_Map(base, util.To_Map(crowd))
	}

	grants, err := client.WhcGetGrants(pid)
	if err == nil {
		base = util.Merge_Map(base, util.To_Map(grants))
	}

	util.FixDecimal(base, "totaltokens")
	util.FixDecimal(base, "tokensissued")
	util.FixDecimal(base, "tokensperunit")

	data, _ := json.Marshal(base)
	//build && upsert
	sp := &common.SmartProperty{PropertyData: string(data), Protocol: common.Wormhole, PropertyID: int64(pid), Issuer: p.Issuer, Ecosystem: common.Production, CreateTxID: createTxId, LastTxID: lastTxId, PropertyName: p.Name, Precision: p.Precision, PropertyCategory: p.Category, PropertySubcategory: p.Subcategory}
	ph := &common.PropertyHistory{PropertyID: int64(pid), TxID: int64(lastTxId)}
	return sp, ph, nil
}
func patchForCrowd(crowd *btcjson.WhcGetCrowdSaleResult, base map[string]interface{}) error {
	txes := crowd.ParticipantTxs
	var total decimal.Decimal
	for _, tx := range txes {
		tokens, err := decimal.NewFromString(tx.ParticipantTokens)
		if err != nil {
			return err
		}

		total = total.Add(tokens)
	}

	tokens, _ := total.Float64()
	base["purchasedtokens"] = tokens
	crowd.ParticipantTxs = nil

	return nil
}

func GetTxByTxHash(lastHash string, ctx context.Context) int {

	tx := model.GetTxByTxHash(lastHash, ctx)
	if tx == nil {
		return -1
	}

	return tx.TxID
}

type FrozenMangeTokenHandler struct{}

func (h *FrozenMangeTokenHandler) Invoke(t *btcjson.GenerateTransactionResult, txId int, amount *decimal.Decimal, ctx context.Context) (*common.AddressesInTx, error) {
	tx := &common.AddressesInTx{Address: t.ReferenceAddress, PropertyID: t.PropertyID, Protocol: common.Wormhole, TxID: txId, AddressRole: common.Issuer, AddressTxIndex: 1}

	//Insert Property History
	insertPropertyHistory(t.PropertyID, txId, ctx)
	return tx, nil
}
