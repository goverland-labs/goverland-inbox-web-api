package chain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"

	"github.com/goverland-labs/inbox-web-api/internal/config"
)

type ChainID int

type Info struct {
	ID               ChainID `json:"id"`
	Name             string  `json:"name"`
	Balance          float64 `json:"balance"`
	Symbol           string  `json:"symbol"`
	FeeApproximation float64 `json:"fee_approximation"`
	TxScanTemplate   string  `json:"tx_scan_template"`
}

type chainInstance struct {
	chain          string
	client         *ethclient.Client
	chainID        ChainID
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

type EstimateParams struct {
	From common.Address
	To   *common.Address
	Data []byte
}

func NewService(cfg config.Chain) (*Service, error) {
	chainInstances := []config.ChainInstance{
		cfg.Eth,
		cfg.Gnosis,
	}

	chains := make(map[ChainID]chainInstance)
	for _, ci := range chainInstances {
		chClient, err := ethclient.Dial(ci.PublicNode)
		if err != nil {
			return nil, err
		}
		chainID := ChainID(ci.ID)

		chains[chainID] = chainInstance{
			chain:          ci.InternalName,
			chainID:        chainID,
			client:         chClient,
			publicName:     ci.PublicName,
			symbol:         ci.Symbol,
			txScanTemplate: ci.TxScanTemplate,
			decimals:       int32(ci.Decimals),
		}
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

func (s *Service) GetChainsInfo(address common.Address) (map[string]Info, error) {
	info := make(map[string]Info)

	for _, chain := range s.chains {
		balance, err := chain.client.BalanceAt(context.Background(), address, nil)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to get balance for chain %s: %w", ErrChainRequestUnreachable, chain.chain, err)
		}

		balanceDec := decimal.NewFromBigInt(balance, -chain.decimals)

		gasPrice, err := chain.client.SuggestGasPrice(context.Background())
		if err != nil {
			return nil, fmt.Errorf("%w: failed to get gas price for chain %s: %w", ErrChainRequestUnreachable, chain.chain, err)
		}

		fee := decimal.NewFromBigInt(gasPrice, -chain.decimals).Mul(decimal.NewFromInt(250000))

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
		return TxStatusWrapper{}, fmt.Errorf("%w: failed to get transaction by hash: %w", ErrChainRequestUnreachable, err)
	}

	if isPending {
		return TxStatusWrapper{
			Status: TxStatusPending,
		}, nil
	}

	receipt, err := chain.client.TransactionReceipt(ctx, txHash)
	if err != nil {
		return TxStatusWrapper{}, fmt.Errorf("%w: failed to get transaction receipt: %w", ErrChainRequestUnreachable, err)
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
	// TODO: config?
	return "0xDE1e8A7E184Babd9F0E3af18f40634e9Ed6F0905"
}

func (s *Service) GetGasPriceHex(chainID ChainID) (string, error) {
	chain, ok := s.chains[chainID]
	if !ok {
		return "", fmt.Errorf("chain with id %d not found", chainID)
	}

	gasPrice, err := chain.client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("%w: failed to get gas price for chain %s: %w", ErrChainRequestUnreachable, chain.chain, err)
	}

	return fmt.Sprintf("0x%x", gasPrice), nil
}

func (s *Service) GetMaxPriorityFeePerGasHex(chainID ChainID) (string, error) {
	chain, ok := s.chains[chainID]
	if !ok {
		return "", fmt.Errorf("chain with id %d not found", chainID)
	}

	gasTipCap, err := chain.client.SuggestGasTipCap(context.Background())
	if err != nil {
		return "", fmt.Errorf("%w: failed to get gas tip cap for chain %s: %w", ErrChainRequestUnreachable, chain.chain, err)
	}

	return fmt.Sprintf("0x%x", gasTipCap), nil
}

func (s *Service) GetGasLimitForSetDelegatesHex(chainID ChainID, params EstimateParams) (string, error) {
	chain, ok := s.chains[chainID]
	if !ok {
		return "", fmt.Errorf("chain with id %d not found", chainID)
	}

	gasLimit, err := chain.client.EstimateGas(context.Background(), ethereum.CallMsg{
		From: params.From,
		To:   params.To,
		Data: params.Data,
	})
	if err != nil {
		return "", fmt.Errorf("%w: failed to estimate gas for chain %s: %w", ErrChainRequestUnreachable, chain.chain, err)
	}

	return fmt.Sprintf("0x%x", gasLimit), nil
}

func (s *Service) SetDelegationABIPack(dao string, delegation []Delegation, expirationTimestamp *big.Int) ([]byte, error) {
	input, err := s.splitDelegationABI.Pack("setDelegation", dao, delegation, expirationTimestamp)
	if err != nil {
		return nil, err
	}

	return input, nil
}
