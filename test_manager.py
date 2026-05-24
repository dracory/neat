import sys
import os
import subprocess

def run_command(command):
    print(f"Running: {command}")
    result = subprocess.run(command, shell=True, capture_output=True, text=True)
    return result

def main():
    if len(sys.argv) < 2:
        print("Usage: python3 test_manager.py <file_path>")
        sys.exit(1)

    file_path = sys.argv[1]
    if not os.path.exists(file_path):
        print(f"File not found: {file_path}")
        sys.exit(1)

    # Enable the test
    with open(file_path, 'r') as f:
        content = f.read()

    if '//go:build disabled' in content:
        new_content = content.replace('//go:build disabled', '//go:build integration')
        with open(file_path, 'w') as f:
            f.write(new_content)
        print(f"Enabled {file_path}")
    elif '//go:build integration' not in content:
        # Some files might not have the tag at all, but we are looking for the ones we disabled
        pass

    # Run the test
    dir_path = os.path.dirname(file_path)
    # Get the file name to run only tests in this file if possible,
    # but go test usually runs the whole package.
    # To be safe and specific, we run the package with integration tag.
    result = run_command(f"go test -v -tags=integration ./{dir_path}")

    if result.returncode == 0:
        print(f"Tests passed for {file_path}")
        # Commit and push
        file_name = os.path.basename(file_path)
        run_command("git add .")
        commit_result = run_command(f'git commit -m "Enable and fix integration test: {file_name}"')
        if commit_result.returncode == 0:
            push_result = run_command("git push origin jules-4636329396559071194-fcebbc3b")
            if push_result.returncode == 0:
                print(f"Successfully pushed {file_name}")
            else:
                print(f"Failed to push {file_name}: {push_result.stderr}")
                sys.exit(1)
        else:
            # Maybe nothing to commit if it already passed before
            if "nothing to commit" in commit_result.stdout:
                 print(f"Nothing to commit for {file_name}")
            else:
                print(f"Failed to commit {file_name}: {commit_result.stdout} {commit_result.stderr}")
                sys.exit(1)
    else:
        print(f"Tests failed for {file_path}")
        print(result.stdout)
        print(result.stderr)
        # We don't revert the tag here so Jules can fix it
        sys.exit(1)

if __name__ == "__main__":
    main()
