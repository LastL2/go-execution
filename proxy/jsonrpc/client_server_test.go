package jsonrpc_test

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/rollkit/go-execution/mocks"
	jsonrpcproxy "github.com/rollkit/go-execution/proxy/jsonrpc"
	"github.com/rollkit/go-execution/types"
)

func TestClientServer(t *testing.T) {
	mockExec := mocks.NewMockExecutor(t)
	config := &jsonrpcproxy.Config{
		DefaultTimeout: 5 * time.Second,
		MaxRequestSize: 1024 * 1024,
	}
	server := jsonrpcproxy.NewServer(mockExec, config)

	testServer := httptest.NewServer(server)
	defer testServer.Close()

	client := jsonrpcproxy.NewClient()
	client.SetConfig(config)

	err := client.Start(testServer.URL)
	require.NoError(t, err)
	defer func() { _ = client.Stop() }()

	t.Run("InitChain", func(t *testing.T) {
		genesisTime := time.Now().UTC().Truncate(time.Second)
		initialHeight := uint64(1)
		chainID := "test-chain"

		expectedStateRoot := make([]byte, 32)
		copy(expectedStateRoot, []byte{1, 2, 3})
		var stateRootHash types.Hash
		copy(stateRootHash[:], expectedStateRoot)

		expectedMaxBytes := uint64(1000000)

		// convert time to Unix and back to ensure consistency
		unixTime := genesisTime.Unix()
		expectedTime := time.Unix(unixTime, 0).UTC()

		mockExec.On("InitChain", mock.Anything, expectedTime, initialHeight, chainID).
			Return(stateRootHash, expectedMaxBytes, nil).Once()

		stateRoot, maxBytes, err := client.InitChain(context.TODO(), genesisTime, initialHeight, chainID)

		require.NoError(t, err)
		assert.Equal(t, stateRootHash, stateRoot)
		assert.Equal(t, expectedMaxBytes, maxBytes)
		mockExec.AssertExpectations(t)
	})

	t.Run("GetTxs", func(t *testing.T) {
		expectedTxs := []types.Tx{[]byte("tx1"), []byte("tx2")}
		mockExec.On("GetTxs", mock.Anything).Return(expectedTxs, nil).Once()

		txs, err := client.GetTxs(context.TODO())
		require.NoError(t, err)
		assert.Equal(t, expectedTxs, txs)
		mockExec.AssertExpectations(t)
	})

	t.Run("ExecuteTxs", func(t *testing.T) {
		txs := []types.Tx{[]byte("tx1"), []byte("tx2")}
		blockHeight := uint64(1)
		timestamp := time.Now().UTC().Truncate(time.Second)

		var prevStateRoot types.Hash
		copy(prevStateRoot[:], []byte{1, 2, 3})

		var expectedStateRoot types.Hash
		copy(expectedStateRoot[:], []byte{4, 5, 6})

		expectedMaxBytes := uint64(1000000)

		// convert time to Unix and back to ensure consistency
		unixTime := timestamp.Unix()
		expectedTime := time.Unix(unixTime, 0).UTC()

		mockExec.On("ExecuteTxs", mock.Anything, txs, blockHeight, expectedTime, prevStateRoot).
			Return(expectedStateRoot, expectedMaxBytes, nil).Once()

		updatedStateRoot, maxBytes, err := client.ExecuteTxs(context.TODO(), txs, blockHeight, timestamp, prevStateRoot)

		require.NoError(t, err)
		assert.Equal(t, expectedStateRoot, updatedStateRoot)
		assert.Equal(t, expectedMaxBytes, maxBytes)
		mockExec.AssertExpectations(t)
	})

	t.Run("SetFinal", func(t *testing.T) {
		blockHeight := uint64(1)
		mockExec.On("SetFinal", mock.Anything, blockHeight).Return(nil).Once()

		err := client.SetFinal(context.TODO(), blockHeight)
		require.NoError(t, err)
		mockExec.AssertExpectations(t)
	})
}
