package node

import (
	"context"

	"encoding/hex"
	"net"
	"sync"
	"time"

	"github.com/vazj/blocker/crypto"
	"github.com/vazj/blocker/proto"
	"github.com/vazj/blocker/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

const blockTime = time.Second * 5

type Mempool struct {
	lock sync.RWMutex
	txx  map[string]*proto.Transaction
}

func NewMempool() *Mempool {
	return &Mempool{
		txx: make(map[string]*proto.Transaction),
	}
}

func (m *Mempool) Clear() []*proto.Transaction {
	m.lock.Lock()
	defer m.lock.Unlock()
	txs := make([]*proto.Transaction, 0, len(m.txx))

	it := 0
	for k, v := range m.txx {
		txs = append(txs, v)
		delete(m.txx, k)
		txs[it] = v
		it++
	}
	//maps.Clear(m.txx)
	//m.txx = make(map[string]*proto.Transaction)

	for _, tx := range m.txx {
		txs = append(txs, tx)
	}

	return txs
}

func (m *Mempool) Len() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return len(m.txx)
}

func (m *Mempool) Has(tx *proto.Transaction) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, ok := m.txx[hex.EncodeToString(types.HashTransaction(tx))]
	return ok
}

func (m *Mempool) Add(tx *proto.Transaction) bool {
	if m.Has(tx) {
		return false
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	m.txx[hex.EncodeToString(types.HashTransaction(tx))] = tx
	return true
}

type ServerConfig struct {
	Version    string
	ListenAddr string
	PrivateKey *crypto.PrivateKey
}

type Node struct {
	ServerConfig
	logger *zap.SugaredLogger

	peerLock sync.RWMutex
	peers    map[proto.NodeClient]*proto.Version
	mempool  *Mempool

	proto.UnimplementedNodeServer
}

func NewNode(cfg ServerConfig) *Node {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.DisableCaller = true
	loggerConfig.Level.SetLevel(zap.DebugLevel)
	logger, _ := loggerConfig.Build()
	return &Node{
		peers:        make(map[proto.NodeClient]*proto.Version),
		logger:       logger.Sugar(),
		mempool:      NewMempool(),
		ServerConfig: cfg,
	}
}

func (n *Node) Start(listenAddr string, boostrapnodes []string) error {
	n.ListenAddr = listenAddr
	var (
		opts       = []grpc.ServerOption{}
		grpcServer = grpc.NewServer(opts...)
	)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		n.logger.Fatal(err)
	}
	proto.RegisterNodeServer(grpcServer, n)
	n.logger.Infow("node started...", "port", n.ListenAddr)

	// bootstrap network with a list of already know nodes
	if len(boostrapnodes) > 0 {
		go n.bootstrapNetwork(boostrapnodes)
	}

	if n.PrivateKey != nil {
		go n.validatorLoop()
	}

	return grpcServer.Serve(ln)
}

func (n *Node) GetVersion() string {
	return n.Version
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	c, err := makeNodeClient(v.ListenAddr)
	if err != nil {
		return nil, err
	}
	n.addPeer(c, v)

	return n.getVersion(), nil
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	if n.mempool.Add(tx) {
		peer, _ := peer.FromContext(ctx)
		hash := hex.EncodeToString(types.HashTransaction(tx))
		n.logger.Debugw("received transaction", "from", peer.Addr, "hash", hash, "we", n.ListenAddr)
		go func() {
			if err := n.broadcast(tx); err != nil {
				n.logger.Errorw("error broadcasting transaction", "err", err)
			}
		}()
	}

	return &proto.Ack{}, nil
}

func (n *Node) validatorLoop() {
	n.logger.Infow("starting validator loop...", "pubKey", n.PrivateKey.PublicKey, "blocktime", blockTime)
	ticker := time.NewTicker(blockTime)
	for {
		select {
		case <-ticker.C:
			txs := n.mempool.Clear()
			n.logger.Debug("time to create a new block", "lenTx", len(txs))
		}
	}
}

func (n *Node) broadcast(msg any) error {
	n.peerLock.RLock()
	defer n.peerLock.RUnlock()

	for p := range n.peers {
		switch v := msg.(type) {
		case *proto.Transaction:
			_, err := p.HandleTransaction(context.Background(), v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (n *Node) bootstrapNetwork(addrs []string) error {
	for _, addr := range addrs {
		if !n.canConnectWith(addr) {
			continue
		}
		n.logger.Debugw("dialing remote node",
			"we", n.ListenAddr,
			"remote node", addr)
		c, v, err := n.dialRemoteNode(addr)
		if err != nil {
			return err
		}
		n.addPeer(c, v)
	}
	return nil
}

func (n *Node) dialRemoteNode(addr string) (proto.NodeClient, *proto.Version, error) {
	c, err := makeNodeClient(addr)
	if err != nil {
		return nil, nil, err
	}
	v, err := c.Handshake(context.Background(), n.getVersion())
	if err != nil {
		return nil, nil, err
	}
	return c, v, nil
}

func makeNodeClient(addr string) (proto.NodeClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return proto.NewNodeClient(conn), nil
}

func (n *Node) addPeer(c proto.NodeClient, v *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()

	// TODO we need to decide if we accept or reject the peer
	n.peers[c] = v

	if len(v.PeerList) > 0 {
		go n.bootstrapNetwork(v.PeerList)
	}

	n.logger.Infow("peer added", "peer", v)
}

func (n *Node) deletePeer(c proto.NodeClient) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	delete(n.peers, c)
}

func (n *Node) getVersion() *proto.Version {
	return &proto.Version{
		Version:    n.Version,
		Height:     0,
		ListenAddr: n.ListenAddr,
		PeerList:   n.getPeerList(),
	}
}

func (n *Node) canConnectWith(addr string) bool {
	if n.ListenAddr == addr {
		return false
	}
	connectedPeers := n.getPeerList()
	for _, connectedAddr := range connectedPeers {
		if addr == connectedAddr {
			return false
		}
	}
	return true
}

func (n *Node) getPeerList() []string {
	n.peerLock.RLock()
	defer n.peerLock.RUnlock()

	addrs := make([]string, 0, len(n.peers))
	for _, v := range n.peers {
		addrs = append(addrs, v.ListenAddr)
	}
	return addrs
}
