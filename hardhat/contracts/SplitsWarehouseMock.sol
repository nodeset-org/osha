// SPDX-License-Identifier: GPL-3.0-or-later
pragma solidity ^0.8.23;

contract SplitsWarehouseMock{
    mapping(address owner => mapping(uint256 id => uint256 amount)) public balanceOf;
    address public constant NATIVE_TOKEN = 0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE;

    function withdraw(
        address _owner,
        address[] calldata _tokens,
        uint256[] calldata _amounts,
        address _withdrawer
    )
        external
    {
        
    }

    function update(address _owner, address _token) external {
        balanceOf[_owner][toUint256(_token)] = 100;
    }

    function toUint256(address _value) internal pure returns (uint256) {
        return uint256(uint160(_value));
    }
}