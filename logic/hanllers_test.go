package engine

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/bcext/gcash/chaincfg/chainhash"
	"github.com/copernet/whccommon/log"
	common "github.com/copernet/whccommon/model"
	"github.com/copernet/whcengine/model"
	"github.com/jinzhu/gorm"
)

/**
txtype=0,SimpleSendHandller
*/
func TestSimpleSendHandller_Invoke(t *testing.T) {
	list := make([]string, 10)
	//list[0] = "90cba742375a9470cca8df55228dec2ef68a11b7066981444d1fec5e3f16c5a1" //tx_type=0,valid,precision=8
	//list[1] = "1392d580e1e9fd48984100d4efe69ad78ed9cd26cbd133f07e776f6b7809f6cc" //tx_type=0,valid,precision=2
	list[0] = "b4135bf01b1e3566fabc65cae82f9504c1f4cbc64e0d189b2169427d829cae2b" //tx_type=0,invalid

	invoke(list)
}

/**
txtype=1,BuyCrowdsaleToken
*/
func TestBuyCrowdTokenHandller_Invoke(t *testing.T) {
	list := make([]string, 2)
	list[0] = "7ff9640d162296c10c7616deb802635fc9e36a79970e4ca2ec8ce31ab56c4086" //tx_type=1,valid
	//list[0] = "7447747c8f22b06c2c7a7cdf0b5e19444179bad347e1ac99f90376f63e875ecd" //tx_type=1,invalid

	invoke(list)
}

/**
txtype=3,SendToOwners
*/
func TestSendToOwnersHandler_Invoke(t *testing.T) {
	list := make([]string, 2)
	//list[0] = "e24161624fe5de54f99d9bdb3ff7d6ebf1c3040cc61f5e7f31dbf45eb69fb1d1" //tx_type=3,valid,1 recipients
	list[1] = "bf3d30fc9c9424bdc6e38fc55320bad6cda9488e74296fc8dfb06cb2d9ee0fd9" //tx_type=3,valid,86 recipients
	//list[0] = "2027078de2eaeeec305e783c8b71725fba29a92d0c8997db0391a8ff7b3cc2b4" //tx_type=3,invalid

	invoke(list)
}

/**
txtype=4,SendAll
*/
func TestSendAllHandler_Invoke(t *testing.T) {
	list := make([]string, 2)
	//list[0] = "31561362b8ed7c502f5f8ccc1358f6292ecb90a743d87315140821893272e7c6" //tx_type=4,valid
	list[0] = "569a0b661686a34639f332af4534c6f7adbc92ddf8aaca40bdbc57ba0bebc8af" //tx_type=4,invalid

	invoke(list)
}

/**
txtype=50,CreateFixedToken
*/
func TestCreateFixedTokenHandler_Invoke(t *testing.T) {
	list := make([]string, 2)
	list[0] = "c4f1932b32cb79b3fb1685b26af021c86d06b0ca18c9674460899be7a8a3e671" //tx_type=50,valid
	//list[0] = "6f8fc3a7d4dd37fdacb31ca61cff90ebdd5d42b00c00712ee39bb45d10e3b340" //tx_type=50,invalid

	invoke(list)
}

/**
txtype=51,CreateCrowdToken
*/
func TestCreateCrowdTokenHandler_Invoke(t *testing.T) {
	list := make([]string, 2)
	list[0] = "73301b2f722490d6c2182936d279968cd247eb36fcfc58ef03921bd25e6744b5" //tx_type=51,valid
	//list[0] = "e957f26ce142f5b00dc3e1fe648fe13198bdbb3b3cde57ecdb8278169993a2a3" //tx_type=51,invalid

	invoke(list)
}

/**
txtype=53,CloseCrowdToken
*/
func TestCloseCrowdTokenHandler_Invoke(t *testing.T) {
	list := make([]string, 2)
	list[0] = "3d0bd525f3ca644fbba3182c0ab84aa9e7880e24283bb67ac3de789331bc2ec7" //tx_type=53,valid
	//list[0] = "c47666604499a681c0a07917696aca3497d59ae87fe8d5b441a4737b56a90af7" //tx_type=53,invalid

	invoke(list)
}

/**
txtype=54,CreateMangeToken
*/
func TestCreateMangeTokenHandler_Invoke(t *testing.T) {
	list := make([]string, 1)
	list[0] = "6579dea76c3d0b4463671c5476f90f20c746992c300a4b8ec4ce6748c0960836" //tx_type=54,valid
	//list[0] = "28c1f4ebc806d5b82b4e877aeeec9cd42637add2d66e8a262c0f0b76f294a77f" //tx_type=54,invalid

	invoke(list)
}

/**
txtype=55,GrantToken
*/
func TestGrantTokenHandler_Invoke(t *testing.T) {
	list := make([]string, 1)
	//list[0] = "f54fe3c9bc2529e39160ce29fcb738c2ec3ea8d265af56152258af1f50b12139" //tx_type=55,valid sendingaddress == referenceaddress
	//list[0] = "64fcc93e20e479140b1cdd719971eca26905d39a75284d4d0d4ff009b2c4a29b" //tx_type=55,valid sendingaddress != referenceaddress
	list[0] = "662a6e26d693130546b214ad3a0e272a2a7661b908d57ddb8bbabc98019dcdc6" //tx_type=55,invalid

	invoke(list)
}

/**
txtype=56,RevokeToken
*/
func TestRevokeTokenHandler_Invoke(t *testing.T) {

	list := make([]string, 2)
	//list[0] = "e41c5a52e4985d07308d98d029c92c5350f1169dfd7af1a5ab20b956c90f8c86" //tx_type=56,valid
	list[0] = "094d9176c2b5418e2a354bf2bd617072bb4682c12576ccc43d27160bbb118cdf" //tx_type=56,invalid

	invoke(list)
}

/**
txtype=68,BurnBCH
*/
func TestBurnBCHHandller_Invoke(t *testing.T) {
	list := make([]string, 1)
	//list[0] = "47cbaa917b366533d98c13f0daa4780e13d8f351eb500930b13100245a0a3045" //tx_type=68,valid
	list[0] = "8489e8c7c88d46fe152e9f6069047a6fed307721211af65d63a0d5134da15b4d" //tx_type=68,valid,unmature
	//list[0] = "ae22050b5c7ce32093cf3d0f7913172efb7c5cdc73b19e7e3502393a407adbbf" //tx_type=68,invalid

	invoke(list)
}

/**
txtype=70,ChangeTokenIssuer
*/
func TestChangeTokenIssuerHandler_Invoke(t *testing.T) {
	list := make([]string, 1)
	list[0] = "3b80af0a138c3c1eeb224cb9ea9893d7627f770eced48c5e30a9ffa144960a06" //tx_type=70,valid
	//list[0] = "13f22e5a39ec41a9d9ae32a63965b9c237eb5cdf27dae586299ff5f227e06da7" //tx_type=70,invalid

	invoke(list)
}

/**
txtype=185,ChangeTokenIssuer
*/
func TestFrozenMangeTokenHandler_Invoke(t *testing.T) {
	list := make([]string, 1)
	list[0] = "7e50a7d3eec2262615996c2a407158677e56d14d656e9ea1711c9a55903ba24d" //tx_type=70,valid
	//list[0] = "13f22e5a39ec41a9d9ae32a63965b9c237eb5cdf27dae586299ff5f227e06da7" //tx_type=70,invalid

	invoke(list)
}

func invoke(list []string) {
	lastTx := model.GetLastTx(model.Asc, log.NewContext())
	var lastTxId int
	if lastTx != nil {
		lastTxId = lastTx.TxID
	}

	x := 1
	for _, tx := range list {

		if tx == "" {
			continue
		}

		transaction, err := client.WhcGetTransaction(tx)
		if err != nil {
			fmt.Println(err.Error())
		}

		txs := model.BeginTransaction()
		ctx := context.WithValue(log.NewContext(), model.TRANSACTION, txs)
		lastTxId += 1

		hash, _ := chainhash.NewHashFromStr(tx)
		txret, err := client.GetRawTransaction(hash)
		buf := bytes.NewBuffer(make([]byte, 0, txret.MsgTx().SerializeSize()))
		txret.MsgTx().Serialize(buf)
		rawdata := hex.EncodeToString(buf.Bytes())

		model.InsertTx(transaction, lastTxId, x, ctx)
		err = insertTxAddr(transaction, lastTxId, ctx)
		model.InsertTxJson(transaction, &rawdata, lastTxId, ctx)

		if err != nil {
			txs = ctx.Value(model.TRANSACTION).(*gorm.DB)
			txs.Rollback()
		}

		//commit
		txs.Commit()
		x += 1
	}

}

func TestGetAll(t *testing.T) {
	result, err := client.WhcGetAllBalancesForAddress("bchtest:qz7y04xzmdqrhrv6dd54klnnpdeej4dy2ctge0yq5n")
	fmt.Println(err)
	fmt.Println(result)

	balance := model.GetAddressBalance("bchtest:qzjtnzcvzxx7s0na88yrg3zl28wwvfp97538sgrrmr", common.Wormhole, 1, nil)
	res, _ := json.Marshal(balance)
	fmt.Println(string(res))
}

func TestPublish(t *testing.T) {
	//factory.Publish(model.UpdateBlockTip, "xxx", log.NewContext())

	val := getAmount("129.02513912")
	fmt.Println(val)

}
