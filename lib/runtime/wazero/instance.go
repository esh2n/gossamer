package wazero_runtime

import (
	"context"
	"fmt"

	"github.com/ChainSafe/gossamer/internal/log"
	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/lib/crypto"
	"github.com/ChainSafe/gossamer/lib/keystore"
	"github.com/ChainSafe/gossamer/lib/runtime"
	"github.com/ChainSafe/gossamer/lib/runtime/offchain"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// Name represents the name of the interpreter
const Name = "wazero"

// Instance backed by wazero.Runtime
type Instance struct {
	Runtime wazero.Runtime
	Module  api.Module
	// Allocator *runtime.FreeingBumpHeapAllocator
	Context *runtime.Context
}

// Config is the configuration used to create a Wasmer runtime instance.
type Config struct {
	Storage     runtime.Storage
	Keystore    *keystore.GlobalKeystore
	LogLvl      log.Level
	Role        common.NetworkRole
	NodeStorage runtime.NodeStorage
	Network     runtime.BasicNetwork
	Transaction runtime.TransactionState
	CodeHash    common.Hash
}

// NewInstance instantiates a runtime from raw wasm bytecode
func NewInstance(code []byte, cfg Config) (instance *Instance, err error) {
	ctx := context.Background()
	rt := wazero.NewRuntime(ctx)

	_, err = rt.NewHostModuleBuilder("env").
		// values from newer kusama/polkadot runtimes
		ExportMemory("memory", 23).
		NewFunctionBuilder().
		WithFunc(ext_logging_log_version_1).
		Export("ext_logging_log_version_1").
		NewFunctionBuilder().
		WithFunc(ext_crypto_ed25519_generate_version_1).
		Export("ext_crypto_ed25519_generate_version_1").
		NewFunctionBuilder().
		WithFunc(func() int32 {
			return 0
		}).
		Export("ext_logging_max_level_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32, b int32, c int32) {
			return
		}).
		Export("ext_transaction_index_index_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32, b int32) {
			return
		}).
		Export("ext_transaction_index_renew_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32) {
			return
		}).
		Export("ext_sandbox_instance_teardown_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32, b int64, c int64, d int32) int32 {
			return 0
		}).
		Export("ext_sandbox_instantiate_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32, b int64, c int64, d int32, e int32, f int32) int32 {
			return 0
		}).
		Export("ext_sandbox_invoke_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32, b int32, c int32, d int32) int32 {
			return 0
		}).
		Export("ext_sandbox_memory_get_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32, b int32, c int32, d int32) int32 {
			return 0
		}).
		Export("ext_sandbox_memory_set_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32) {
			return
		}).
		Export("ext_sandbox_memory_teardown_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32, b int64) int32 {
			return 0
		}).
		Export("ext_crypto_ed25519_generate_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32) int64 {
			return 0
		}).
		Export("ext_crypto_ed25519_public_keys_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32, b int32, c int64) int64 {
			return 0
		}).
		Export("ext_crypto_ed25519_sign_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32, b int64, c int32) int32 {
			return 0
		}).
		Export("ext_crypto_ed25519_verify_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32, b int32) int64 {
			return 0
		}).
		Export("ext_crypto_secp256k1_ecdsa_recover_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32, b int32) int64 {
			return 0
		}).
		Export("ext_crypto_secp256k1_ecdsa_recover_version_2").
		NewFunctionBuilder().
		WithFunc(func(a int32, b int64, c int32) int32 {
			return 0
		}).
		Export("ext_crypto_ecdsa_verify_version_2").
		NewFunctionBuilder().
		WithFunc(func(a int32, b int32) int64 {
			return 0
		}).
		Export("ext_crypto_secp256k1_ecdsa_recover_compressed_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32, b int32) int64 {
			return 0
		}).
		Export("ext_crypto_secp256k1_ecdsa_recover_compressed_version_2").
		NewFunctionBuilder().
		WithFunc(func(a int32, b int64) int32 {
			return 0
		}).
		Export("ext_crypto_sr25519_generate_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32) int64 {
			return 0
		}).
		Export("ext_crypto_sr25519_public_keys_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32, b int32, c int64) int64 {
			return 0
		}).
		Export("ext_crypto_sr25519_sign_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32, b int64, c int32) int32 {
			return 0
		}).
		Export("ext_crypto_sr25519_verify_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32, b int64, c int32) int32 {
			return 0
		}).
		Export("ext_crypto_sr25519_verify_version_2").
		NewFunctionBuilder().
		WithFunc(func() {
			return
		}).
		Export("ext_crypto_start_batch_verify_version_1").
		NewFunctionBuilder().
		WithFunc(func() int32 {
			return 0
		}).
		Export("ext_crypto_finish_batch_verify_version_1").
		NewFunctionBuilder().
		WithFunc(func() int32 {
			return 0
		}).
		Export("ext_trie_blake2_256_root_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64) int32 {
			return 0
		}).
		Export("ext_trie_blake2_256_ordered_root_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64, _ int32) int32 {
			return 0
		}).
		Export("ext_trie_blake2_256_ordered_root_version_2").
		NewFunctionBuilder().
		WithFunc(func(a int32, b int64, c int64, d int64) int32 {
			return 0
		}).
		Export("ext_trie_blake2_256_verify_proof_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64) {
			return
		}).
		Export("ext_misc_print_hex_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64) {
			return
		}).
		Export("ext_misc_print_num_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64) {
			return
		}).
		Export("ext_misc_print_utf8_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64) int64 {
			return 0
		}).
		Export("ext_misc_runtime_version_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64, b int64, c int64, d int32) int64 {
			return 0
		}).
		Export("ext_default_child_storage_read_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64, b int64) {
			return
		}).
		Export("ext_default_child_storage_clear_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64, b int64) {
			return
		}).
		Export("ext_default_child_storage_clear_prefix_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64, b int64) int32 {
			return 0
		}).
		Export("ext_default_child_storage_exists_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64, b int64) int64 {
			return 0
		}).
		Export("ext_default_child_storage_get_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64, b int64) int64 {
			return 0
		}).
		Export("ext_default_child_storage_next_key_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64, b int64) int64 {
			return 0
		}).
		Export("ext_default_child_storage_root_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64, b int64, c int64) {
			return
		}).
		Export("ext_default_child_storage_set_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64) {
			return
		}).
		Export("ext_default_child_storage_storage_kill_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64, b int64) int32 {
			return 0
		}).
		Export("ext_default_child_storage_storage_kill_version_2").
		NewFunctionBuilder().
		WithFunc(func(a int64, b int64) int64 {
			return 0
		}).
		Export("ext_default_child_storage_storage_kill_version_3").
		NewFunctionBuilder().
		WithFunc(ext_allocator_free_version_1).
		Export("ext_allocator_free_version_1").
		NewFunctionBuilder().
		WithFunc(ext_allocator_malloc_version_1).
		Export("ext_allocator_malloc_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64) int32 {
			return 0
		}).
		Export("ext_hashing_blake2_128_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64) int32 {
			return 0
		}).
		Export("ext_hashing_blake2_256_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64) int32 {
			return 0
		}).
		Export("ext_hashing_keccak_256_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64) int32 {
			return 0
		}).
		Export("ext_hashing_sha2_256_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64) int32 {
			return 0
		}).
		Export("ext_hashing_twox_256_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64) int32 {
			return 0
		}).
		Export("ext_hashing_twox_128_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64) int32 {
			return 0
		}).
		Export("ext_hashing_twox_64_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64, b int64) {
			return
		}).
		Export("ext_offchain_index_set_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32, b int64) {
			return
		}).
		Export("ext_offchain_local_storage_clear_version_1").
		NewFunctionBuilder().
		WithFunc(func() int32 {
			return 0
		}).
		Export("ext_offchain_is_validator_version_1").
		NewFunctionBuilder().
		WithFunc(func(_ int32, _ int64, _ int64, _ int64) int32 {
			return 0
		}).
		Export("ext_offchain_local_storage_compare_and_set_version_1").
		NewFunctionBuilder().
		WithFunc(func(_ int32, _ int64) int64 {
			return 0
		}).
		Export("ext_offchain_local_storage_get_version_1").
		NewFunctionBuilder().
		WithFunc(func(_ int32, _ int64, _ int64) {
			return
		}).
		Export("ext_offchain_local_storage_set_version_1").
		NewFunctionBuilder().
		WithFunc(func() int64 {
			return 0
		}).
		Export("ext_offchain_network_state_version_1").
		NewFunctionBuilder().
		WithFunc(func() int32 {
			return 0
		}).
		Export("ext_offchain_random_seed_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64) int64 {
			return 0
		}).
		Export("ext_offchain_submit_transaction_version_1").
		NewFunctionBuilder().
		WithFunc(func() int64 {
			return 0
		}).
		Export("ext_offchain_timestamp_version_1").
		NewFunctionBuilder().
		WithFunc(func() int64 {
			return 0
		}).
		Export("ext_offchain_sleep_until_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64, _ int64, c int64) int64 {
			return 0
		}).
		Export("ext_offchain_http_request_start_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64, _ int64, c int64) int64 {
			return 0
		}).
		Export("ext_offchain_http_request_add_header_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64, _ int64) {
			return
		}).
		Export("ext_storage_append_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64, _ int64) {
			return
		}).
		Export("ext_storage_changes_root_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64) {
			return
		}).
		Export("ext_storage_clear_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64) {
			return
		}).
		Export("ext_storage_clear_prefix_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64, _ int64) int64 {
			return 0
		}).
		Export("ext_storage_clear_prefix_version_2").
		NewFunctionBuilder().
		WithFunc(func(a int64) int32 {
			return 0
		}).
		Export("ext_storage_exists_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64) int64 {
			return 0
		}).
		Export("ext_storage_get_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64) int64 {
			return 0
		}).
		Export("ext_storage_next_key_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int64, _ int64, _ int32) int64 {
			return 0
		}).
		Export("ext_storage_read_version_1").
		NewFunctionBuilder().
		WithFunc(func() int64 {
			return 0
		}).
		Export("ext_storage_root_version_1").
		NewFunctionBuilder().
		WithFunc(func(a int32) int64 {
			return 0
		}).
		Export("ext_storage_root_version_2").
		NewFunctionBuilder().
		WithFunc(func(a int64, _ int64) {
			return
		}).
		Export("ext_storage_set_version_1").
		NewFunctionBuilder().
		WithFunc(func() {
			return
		}).
		Export("ext_storage_start_transaction_version_1").
		NewFunctionBuilder().
		WithFunc(func() {
			return
		}).
		Export("ext_storage_rollback_transaction_version_1").
		NewFunctionBuilder().
		WithFunc(func() {
			return
		}).
		Export("ext_storage_commit_transaction_version_1").
		Instantiate(ctx)

	if err != nil {
		panic(err)
	}

	mod, err := rt.Instantiate(ctx, code)
	if err != nil {
		panic(err)
	}

	global := mod.ExportedGlobal("__heap_base")
	if global == nil {
		panic("huh?")
	}
	fmt.Printf("%+v\n", global)
	global.Get()

	hb := api.DecodeU32(global.Get())
	fmt.Println("heapbase", hb)

	mem := mod.Memory()
	if mem == nil {
		panic("wtf?")
	}

	allocator := runtime.NewAllocator(mem, hb)

	return &Instance{
		Runtime: rt,
		Context: &runtime.Context{
			Storage:         cfg.Storage,
			Allocator:       allocator,
			Keystore:        cfg.Keystore,
			Validator:       cfg.Role == common.AuthorityRole,
			NodeStorage:     cfg.NodeStorage,
			Network:         cfg.Network,
			Transaction:     cfg.Transaction,
			SigVerifier:     crypto.NewSignatureVerifier(logger),
			OffchainHTTPSet: offchain.NewHTTPSet(),
		},
		Module: mod,
	}, nil
}
