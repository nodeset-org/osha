async function main() {
    const [deployer] = await ethers.getSigners();

    console.log("Deploying contracts with the account:", deployer.address);

    const balance = await deployer.getBalance();
    console.log("Account balance:", balance.toString());

    const Token = await ethers.getContractFactory("Token");
    const token = await Token.deploy("BlahToken", "BLAH", 1000000);

    console.log("Token deployed at:", token.address);

    const SplitsWarehouseMock = await ethers.getContractFactory("SplitsWarehouseMock");
    const splitsWarehouseMock = await SplitsWarehouseMock.deploy();

    console.log("SplitsWarehouse deployed at:", splitsWarehouseMock.address);
}

main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error);
        process.exit(1);
    });
