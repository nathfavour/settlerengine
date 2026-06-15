import { ethers } from "hardhat";

async function main() {
  const SettlerRegistry = await ethers.getContractFactory("SettlerRegistry");
  const registry = await SettlerRegistry.deploy();

  await registry.waitForDeployment();

  console.log(`SettlerRegistry deployed to: ${await registry.getAddress()}`);
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
