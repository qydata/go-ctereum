// SPDX-License-Identifier: SimPL-2.0
pragma solidity 0.7.6;
pragma abicoder v2;

import "@openzeppelin/contracts-upgradeable/drafts/EIP712Upgradeable.sol";
import "@openzeppelin/contracts-upgradeable/cryptography/ECDSAUpgradeable.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/EnumerableSet.sol";

/**
 * @dev 对以太坊地址实名的实现合约, 通过离线授权或者自己实名
 */
contract AuthController is EIP712Upgradeable, Ownable {
    using AddressUpgradeable for address;
    using ECDSAUpgradeable for bytes32;
    using EnumerableSet for EnumerableSet.AddressSet;

    // 白名单set集合
    EnumerableSet.AddressSet private whitelist;

    string constant SIGNATURE_ERROR = "signature verification error";
    bytes32 public constant AUTH_TYPEHASH =
        keccak256(
            "Authentication(uint256 cid,uint256 uid,address caddress,bool isAuth)"
        );
    // 认证的对象结构体
    struct AuthData {
        uint256 cid;   // 机构标识
        uint256 uid;   // 唯一标识
        address caddress; // 认证地址
        address sender;   // 发起认证操作地址
        bytes signature;  // 签名数据
        bool isAuth;      // 认证状态
    }
    // 订单管理
    mapping(uint256 => bool) public orders;
    // 认证信息管理
    mapping(address => AuthData) public auths;
    // 事件
    // 认证信息事件
    event Authentication(AuthData);
    // 添加白名单事件
    event AddedToWhiteList(address);
    // 移除白名单事件
    event RemovedFromWhiteList(address);

    // 白名单控制
    modifier nonWhitelisted() {
        require(whitelist.contains(msg.sender), "CONTRACT_NOT_WHITELISTED");
        _;
    }

    constructor() public {
        // EIP712对name和版本初始化
        __EIP712_init_unchained("Authentication", "1");
    }

    /**
        通过传入的实名信息, 来对地址进行实名,
        orderId 每个订单号只能使用一次
    */
    function authentication(AuthData calldata auth, uint256 orderId)
        public
        nonWhitelisted
    {
        // 判断订单号是否已经处理
        require(orders[orderId] == false, "order has been processed!");
        // 这里方便对操作人进行索引
        require(auth.sender == msg.sender, "auth sender set err!");

        // 认证地址不是实名方地址, 进行签名验证
        if (msg.sender != auth.caddress) {
            bytes32 hash =
                keccak256(
                    abi.encode(
                        AUTH_TYPEHASH,
                        auth.cid,
                        auth.uid,
                        auth.caddress,
                        auth.isAuth
                    )
                );
            validate(auth.caddress, hash, auth.signature);
        }

        // 数据存储
        auths[auth.caddress] = auth;
        // 修改订单状态
        orders[orderId] = true;
        // 提交事件
        emit Authentication(auth);
    }

    /**
    批量实名
    */
    function authenticationBetch(
        AuthData[] calldata auths,
        uint256[] calldata orderIds
    ) external nonWhitelisted {
        require(auths.length == orderIds.length, "length no metch!");
        for (uint256 i = 0; i < auths.length; i++) {
            authentication(auths[i], orderIds[i]);
        }
    }

    /**
    EIP721 签名验证
    */
    function validate(
        address signer,
        bytes32 structHash,
        bytes memory signature
    ) internal view {
        bytes32 hash = _hashTypedDataV4(structHash);
        require(hash.recover(signature) == signer, SIGNATURE_ERROR);
    }

    /**
    添加白名单
    */
    function addToWhitelist(address[] calldata _addresses) external onlyOwner {
        for (uint256 i = 0; i < _addresses.length; i++) {
            address currentAddress = _addresses[i];
            require(currentAddress != address(0), "ZERO_ADDRESS");
            require(!whitelist.contains(currentAddress), "ALREADY_WHITELISTED");
            whitelist.add(currentAddress);
            emit AddedToWhiteList(currentAddress);
        }
    }

    /**
    移除白名单
    */
    function removeFromWhitelist(address[] calldata _addresses)
        external
        onlyOwner
    {
        for (uint256 i = 0; i < _addresses.length; i++) {
            address currentAddress = _addresses[i];
            require(currentAddress != address(0), "ZERO_ADDRESS");
            require(whitelist.contains(currentAddress), "NOT_WHITELISTED_YET");
            whitelist.remove(currentAddress);
            emit RemovedFromWhiteList(currentAddress);
        }
    }

    /**
    查询地址白名单
    */
    function whitelisted(address _address) external view returns (bool) {
        return whitelist.contains(_address);
    }

    /**
    获得白名单列表
    */
    function getWhitelist() external view returns (address[] memory list) {
        uint256 length = whitelist.length();
        list = new address[](length);
        for (uint256 i = 0; i < length; i++) {
            list[i] = whitelist.at(i);
        }
    }
}
