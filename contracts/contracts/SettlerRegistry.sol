// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract SettlerRegistry {
    address public owner;

    event AgentPaymentLogged(
        bytes32 indexed agentId, 
        uint256 indexed invoiceId, 
        uint256 amount, 
        string metadata
    );

    modifier onlyOwner() {
        require(msg.sender == owner, "Not authorized");
        _;
    }

    constructor() {
        owner = msg.sender;
    }

    // SettlerEngine agents call this to anchor a settled transaction on-chain
    function logAgentPayment(
        bytes32 _agentId, 
        uint256 _invoiceId, 
        uint256 _amount, 
        string calldata _metadata
    ) external {
        emit AgentPaymentLogged(_agentId, _invoiceId, _amount, _metadata);
    }
}
