package chain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
)

const (
	Eth    Chain = "eth"
	Gnosis Chain = "gnosis"

	EthChainID    ChainID = 1
	GnosisChainID ChainID = 100
)

type Chain string
type ChainID int

type Info struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	Balance          float64 `json:"balance"`
	Symbol           string  `json:"symbol"`
	FeeApproximation float64 `json:"fee_approximation"`
	TxScanTemplate   string  `json:"tx_scan_template"`
}

type chainInstance struct {
	chain          Chain
	client         *ethclient.Client
	chainID        int
	publicName     string
	symbol         string
	txScanTemplate string
	decimals       int32
}

const (
	TxStatusPending TxStatus = "pending"
	TxStatusSuccess TxStatus = "success"
	TxStatusFailed  TxStatus = "failed"
)

type TxStatus string

type TxStatusWrapper struct {
	Status TxStatus `json:"status"`
}

type Service struct {
	splitDelegationABI *abi.ABI
	chains             map[ChainID]chainInstance
}

func NewService() (*Service, error) {
	ethClient, err := ethclient.Dial("https://ethereum-rpc.publicnode.com")
	if err != nil {
		return nil, err
	}

	chains := make(map[ChainID]chainInstance, 0)
	chains[EthChainID] = chainInstance{
		chain:          Eth,
		chainID:        1,
		client:         ethClient,
		publicName:     "Ethereum",
		symbol:         "eth",
		txScanTemplate: "https://etherscan.io/tx/:id",
		decimals:       18,
	}

	gnosisClient, err := ethclient.Dial("https://gnosis-rpc.publicnode.com")
	if err != nil {
		return nil, err
	}

	chains[GnosisChainID] = chainInstance{
		chain:          Gnosis,
		client:         gnosisClient,
		chainID:        100,
		publicName:     "Gnosis Chain",
		symbol:         "xDai",
		txScanTemplate: "https://gnosisscan.io/tx/:id",
		decimals:       18,
	}

	sdABI, err := getSplitDelegationAbi()
	if err != nil {
		return nil, err
	}

	return &Service{
		chains:             chains,
		splitDelegationABI: sdABI,
	}, nil
}

func (s *Service) GetChainsInfo(address common.Address) (map[Chain]Info, error) {
	info := make(map[Chain]Info)

	for _, chain := range s.chains {
		balance, err := chain.client.BalanceAt(context.Background(), address, nil)
		if err != nil {
			return nil, err
		}

		balanceDec := decimal.NewFromBigInt(balance, -chain.decimals)

		gasPrice, err := chain.client.SuggestGasPrice(context.Background())
		if err != nil {
			return nil, err
		}

		fee := decimal.NewFromBigInt(gasPrice, -chain.decimals).Mul(decimal.NewFromInt(50000))

		info[chain.chain] = Info{
			ID:               chain.chainID,
			Name:             chain.publicName,
			Balance:          balanceDec.InexactFloat64(),
			Symbol:           chain.symbol,
			FeeApproximation: fee.InexactFloat64(),
			TxScanTemplate:   chain.txScanTemplate,
		}
	}

	return info, nil
}

func (s *Service) GetTxStatus(ctx context.Context, chainID ChainID, txHashHex string) (TxStatusWrapper, error) {
	chain, ok := s.chains[chainID]
	if !ok {
		return TxStatusWrapper{}, fmt.Errorf("chain with id %d not found", chainID)
	}

	txHash := common.HexToHash(txHashHex)

	_, isPending, err := chain.client.TransactionByHash(ctx, txHash)
	if err != nil {
		return TxStatusWrapper{}, err
	}

	if isPending {
		return TxStatusWrapper{
			Status: TxStatusPending,
		}, nil
	}

	receipt, err := chain.client.TransactionReceipt(ctx, txHash)
	if err != nil {
		return TxStatusWrapper{}, err
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		return TxStatusWrapper{
			Status: TxStatusSuccess,
		}, nil
	}

	return TxStatusWrapper{
		Status: TxStatusFailed,
	}, nil
}

func (s *Service) GetDelegatesContractAddress(_ ChainID) string {
	return "0xDE1e8A7E184Babd9F0E3af18f40634e9Ed6F0905"
}

func (s *Service) GetGasPriceHex(chainID ChainID) (string, error) {
	chain, ok := s.chains[chainID]
	if !ok {
		return "", fmt.Errorf("chain with id %d not found", chainID)
	}

	gasPrice, err := chain.client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("0x%x", gasPrice), nil
}

func (s *Service) GetGasLimitForSetDelegatesHex() (string, error) {
	return fmt.Sprintf("0x%x", 50000), nil
}

func (s *Service) SetDelegationABIPack(dao string, delegation []Delegation, expirationTimestamp *big.Int) (string, error) {
	input, err := s.splitDelegationABI.Pack("setDelegation", dao, delegation, expirationTimestamp)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("0x%x", input), nil
}
