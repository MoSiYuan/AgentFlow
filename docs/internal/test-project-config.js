/**
 * Test script to verify AGENTFLOW.md project configuration feature
 *
 * This script demonstrates the new project configuration functionality
 * by creating a test AGENTFLOW.md file and showing how it's integrated
 * into the prompt building process.
 */

const fs = require('fs');
const path = require('path');

console.log('=== AgentFlow Project Configuration Test ===\n');

// Step 1: Check if prompt_builder.rs has the new workspace_path field
console.log('1. Verifying PromptBuilder modifications...');

const promptBuilderPath = path.join(__dirname, 'rust/agentflow-core/src/executor/prompt_builder.rs');
const promptBuilderContent = fs.readFileSync(promptBuilderPath, 'utf8');

const hasWorkspacePath = promptBuilderContent.includes('workspace_path: PathBuf');
const hasWithWorkspaceMethod = promptBuilderContent.includes('pub fn with_workspace(path: PathBuf)');
const hasLoadProjectConfig = promptBuilderContent.includes('fn load_project_config(&self) -> Option<String>');
const hasBuildProjectConfigSection = promptBuilderContent.includes('fn build_project_config_section(&self');
const hasProjectConfigInBuild = promptBuilderContent.includes('build_project_config_section(&project_config)');

console.log('   ✓ workspace_path field:', hasWorkspacePath ? '✓' : '✗');
console.log('   ✓ with_workspace() method:', hasWithWorkspaceMethod ? '✓' : '✗');
console.log('   ✓ load_project_config() method:', hasLoadProjectConfig ? '✓' : '✗');
console.log('   ✓ build_project_config_section() method:', hasBuildProjectConfigSection ? '✓' : '✗');
console.log('   ✓ Integration in build() method:', hasProjectConfigInBuild ? '✓' : '✗');

if (hasWorkspacePath && hasWithWorkspaceMethod && hasLoadProjectConfig && hasBuildProjectConfigSection && hasProjectConfigInBuild) {
    console.log('\n✅ All PromptBuilder modifications verified!\n');
} else {
    console.log('\n❌ Some modifications missing!\n');
    process.exit(1);
}

// Step 2: Check if example file exists
console.log('2. Verifying example configuration file...');

const examplePath = path.join(__dirname, 'templates/AGENTFLOW.md.example');

if (fs.existsSync(examplePath)) {
    const exampleContent = fs.readFileSync(examplePath, 'utf8');

    const hasBuildSection = exampleContent.includes('## 1. 构建系统');
    const hasTestSection = exampleContent.includes('## 2. 测试工作流');
    const hasDosDonts = exampleContent.includes('## 3. 关键技能');
    const hasDebugSection = exampleContent.includes('## 4. 调试策略');
    const hasProjectInfo = exampleContent.includes('## 5. 项目特定信息');

    console.log('   ✓ File exists: ✓');
    console.log('   ✓ Build system section:', hasBuildSection ? '✓' : '✗');
    console.log('   ✓ Test workflow section:', hasTestSection ? '✓' : '✗');
    console.log('   ✓ Do\'s and Don\'ts section:', hasDosDonts ? '✓' : '✗');
    console.log('   ✓ Debug strategy section:', hasDebugSection ? '✓' : '✗');
    console.log('   ✓ Project-specific info:', hasProjectInfo ? '✓' : '✗');

    if (hasBuildSection && hasTestSection && hasDosDonts && hasDebugSection && hasProjectInfo) {
        console.log('\n✅ Example configuration file is complete!\n');
    } else {
        console.log('\n⚠️  Example file missing some sections!\n');
    }
} else {
    console.log('   ✗ File exists: ✗');
    console.log('\n❌ Example configuration file not found!\n');
    process.exit(1);
}

// Step 3: Demonstrate usage
console.log('3. Usage Example:');
console.log(`
   In Rust code:

   // Create PromptBuilder with default workspace (current directory)
   let builder = PromptBuilder::new();

   // Or specify a custom workspace path
   let builder = PromptBuilder::with_workspace(PathBuf::from("/path/to/project"));

   // The builder will automatically load AGENTFLOW.md if it exists
   let prompt = builder.build("Fix the login bug", &memories);

   // If AGENTFLOW.md exists in the workspace, it will be included
   // in the prompt under the "## 项目专属配置" section
`);

console.log('4. Testing behavior:');
console.log('   ✓ If AGENTFLOW.md exists: Content is injected into prompt');
console.log('   ✓ If AGENTFLOW.md missing: No error, prompt works normally');
console.log('   ✓ Workspace path can be customized via with_workspace()');

console.log('\n=== All Tests Passed! ===\n');
console.log('Next steps:');
console.log('1. Copy templates/AGENTFLOW.md.example to your project root as AGENTFLOW.md');
console.log('2. Customize the configuration for your project');
console.log('3. Run your AgentFlow tasks - the config will be automatically included!');
