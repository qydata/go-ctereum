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
            "Authentication(uint256 orderId,address caddress,address sender,bool isAuth)"
        );
    // 认证的对象结构体

    // 字段：
    // TODO
    // 1.实名地址
    // 2.实名机构签名信息
    // 3.实名时间
    // 4.实名有效期
    // 5.实名状态（有效/过期／注销：平台或用户明确发送注销交易）
    // 6.实名等级
    // 7.保留字段（json结构体？）

    struct AuthData {
        address caddress; // 认证地址
        address sender; // 发起认证操作地址
        bytes signature; // 签名数据
        uint256 authTime;
        uint256 authExpiry;
        bool isAuth; // 认证状态
        uint256 authLevel; // 认证状态
        string expandData;
    }
    // 订单管理
    mapping(uint256 => bool) public orders;
    // 认证信息管理
    mapping(address => AuthData) auths;
    // 事件
    // 认证信息事件
    event Authentication(AuthData, address indexed caddress);
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
    function authentication(AuthData memory auth, uint256 orderId)
        public
        nonWhitelisted
    {
        // 判断订单号是否已经处理
        require(orders[orderId] == false, "order has been processed!");
        // 这里方便对操作人进行索引
        require(auth.sender == msg.sender, "auth sender set err!");
        // 判断有效时间
        require(
            block.timestamp < auth.authExpiry,
            "authExpiry must more than currentTIme!"
        );

        auth.authTime = block.timestamp;
        // 认证地址不是实名方地址, 进行签名验证
        if (msg.sender != auth.caddress) {
            bytes32 hash =
                keccak256(
                    abi.encode(
                        AUTH_TYPEHASH,
                        orderId,
                        auth.caddress,
                        auth.sender,
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
        emit Authentication(auth, auth.caddress);
    }

    /**
    方便链上调用
    接收方是合约,将会产生比较复杂的判断问题
     */
    // function authsDouble(address from, address to) external view returns (bool pass) {
    //     pass = auths[from].isAuth && auths[to].isAuth;
    // }
    /**
    实名查询
     如果是合约地址, 则直接是已经实名状态
     */
    function authsSingle(address addr)
        external
        view
        returns (AuthData memory auth)
    {
        if (addr.isContract()) {
            auth = auths[address(0)];
            auth.sender = addr;
            auth.isAuth = true;
        } else {
            auth = auths[addr];
            if (
                auth.sender == address(0) || block.timestamp > auth.authExpiry
            ) {
                auth = auths[addr];
                auth.sender = addr;
                auth.isAuth = false;
            }
        }
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
