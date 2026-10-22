package tunnel

import (
	"fmt"
	"log"
	"net"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Config représente la configuration nécessaire pour monter un tunnel.
type Config struct {
	InterfaceName string
	PrivateKey    string
	ListenPort    int
	PeerPublicKey string
	PeerEndpoint  string
	AllowedIPs    []string
}

// SetupTunnel configure l'interface WireGuard.
func SetupTunnel(cfg Config) error {
	// 1. Se connecter au contrôleur WireGuard (qui parle à l'interface wg0)
	client, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("impossible d'initialiser wgctrl: %w", err)
	}
	defer client.Close()

	// 2. Décoder la clé privée
	privKey, err := wgtypes.ParseKey(cfg.PrivateKey)
	if err != nil {
		return fmt.Errorf("clé privée invalide: %w", err)
	}

	// 3. Préparer la configuration de l'interface
	config := wgtypes.Config{
		PrivateKey:   &privKey,
		ListenPort:   &cfg.ListenPort,
		ReplacePeers: true,
	}

	// 4. Si un pair est spécifié, l'ajouter
	if cfg.PeerPublicKey != "" {
		pubKey, err := wgtypes.ParseKey(cfg.PeerPublicKey)
		if err != nil {
			return fmt.Errorf("clé publique du pair invalide: %w", err)
		}

		peerConfig := wgtypes.PeerConfig{
			PublicKey: pubKey,
		}

		// Si on a une adresse IP pour le pair (ex: 10.0.0.2/32)
		if len(cfg.AllowedIPs) > 0 {
			var allowedIPs []net.IPNet
			for _, ipStr := range cfg.AllowedIPs {
				_, ipNet, err := net.ParseCIDR(ipStr)
				if err != nil {
					return fmt.Errorf("CIDR invalide %s: %w", ipStr, err)
				}
				allowedIPs = append(allowedIPs, *ipNet)
			}
			peerConfig.AllowedIPs = allowedIPs
		}

		// Si on a un endpoint (adresse IP publique et port du pair)
		if cfg.PeerEndpoint != "" {
			// Note: On devrait résoudre l'IP ici, mais pour le test on assume que c'est une IP
			udpAddr, err := net.ResolveUDPAddr("udp", cfg.PeerEndpoint)
			if err != nil {
				log.Printf("Attention: impossible de résoudre l'endpoint %s: %v", cfg.PeerEndpoint, err)
			} else {
				peerConfig.Endpoint = udpAddr
			}
		}

		config.Peers = []wgtypes.PeerConfig{peerConfig}
	}

	// 5. Appliquer la configuration à l'interface
	if err := client.ConfigureDevice(cfg.InterfaceName, config); err != nil {
		return fmt.Errorf("impossible de configurer l'interface %s: %w", cfg.InterfaceName, err)
	}

	log.Printf("Interface %s configurée avec succès.", cfg.InterfaceName)
	return nil
}
