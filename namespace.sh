# !bin/bash

SERVER_NAMESPACE="dp-tcp-server"
CLIENT_NAMESPACE="dp-tcp-client"

usage() {
    echo "Usage: $0 [ up | down | server-ns | client-ns ]"
    echo "  up     - Setup both server and client namespaces"
    echo "  down   - Cleanup both namespaces"
    echo "  server-ns - Enter server namespace"
    echo "  client-ns - Enter client namespace"
    exit 1
}

setup_network_namespace() {
    # Remove exist network namespace
    echo "Removing exist network namespace..."
    cleanup_network_namespace
    echo

    # Create network namespace
    echo "Creating network namespace..."
    sudo ip netns add $SERVER_NAMESPACE 2>/dev/null || true
    sudo ip netns add $CLIENT_NAMESPACE 2>/dev/null || true
    echo

    # Create veth pair
    echo "Creating veth pair..."
    sudo ip link add link1-s type veth peer link1-c
    sudo ip link add link2-s type veth peer link2-c
    echo

    # Add veth pair to network namespace
    echo "Adding veth pair to network namespace..."
    sudo ip link set link1-s netns $SERVER_NAMESPACE
    sudo ip link set link1-c netns $CLIENT_NAMESPACE
    sudo ip link set link2-s netns $SERVER_NAMESPACE
    sudo ip link set link2-c netns $CLIENT_NAMESPACE
    echo

    # Set veth pair up
    echo "Setting veth pair up..."
    sudo ip netns exec $SERVER_NAMESPACE ip link set link1-s up
    sudo ip netns exec $CLIENT_NAMESPACE ip link set link1-c up
    sudo ip netns exec $SERVER_NAMESPACE ip link set link2-s up
    sudo ip netns exec $CLIENT_NAMESPACE ip link set link2-c up
    echo

    # Set veth pair ip
    echo "Setting veth pair ip..."
    sudo ip netns exec $SERVER_NAMESPACE ip addr add 10.0.1.1/24 dev link1-s
    sudo ip netns exec $CLIENT_NAMESPACE ip addr add 10.0.1.2/24 dev link1-c
    sudo ip netns exec $SERVER_NAMESPACE ip addr add 10.0.2.1/24 dev link2-s
    sudo ip netns exec $CLIENT_NAMESPACE ip addr add 10.0.2.2/24 dev link2-c
    echo

    # Set up default route
    echo "Setting up default route..."
    sudo ip netns exec $SERVER_NAMESPACE ip route add default via 10.0.0.1
    sudo ip netns exec $CLIENT_NAMESPACE ip route add default via 10.0.0.2
    echo

    echo "$SERVER_NAMESPACE namespace setup complete"
    echo "$CLIENT_NAMESPACE namespace setup complete"
    echo "Network topology:"
    echo "  $SERVER_NAMESPACE (10.0.1.1) <---> $CLIENT_NAMESPACE (10.0.1.2)"
    echo "  $SERVER_NAMESPACE (10.0.2.1) <---> $CLIENT_NAMESPACE (10.0.2.2)"
}

cleanup_network_namespace() {
    echo "Removing network namespace..."
    sudo ip netns delete $SERVER_NAMESPACE
    sudo ip netns delete $CLIENT_NAMESPACE
    echo "Network namespace removed"
    echo
}

main() {
    if [ $# -ne 1 ]; then
        usage
    fi

    case "$1" in
        "up")
            setup_network_namespace
        ;;
        "down")
            cleanup_network_namespace
        ;;
        "server-ns")
            sudo ip netns exec $SERVER_NAMESPACE bash
        ;;
        "client-ns")
            sudo ip netns exec $CLIENT_NAMESPACE bash
        ;;
        *)
            usage
        ;;
    esac
}

main "$@"