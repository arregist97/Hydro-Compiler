import subprocess
import unittest
import os
import shutil

class TestHydroCompiler(unittest.TestCase):
    def setUp(self):
        # Path to the build directory
        self.build_dir = './../build'

        # Clear the build directory
        if os.path.exists(self.build_dir):
            shutil.rmtree(self.build_dir)  # Remove all files and directories in the build directory
        os.makedirs(self.build_dir)  # Recreate the build directory
    def test_if_statement(self):
        # Paths
        hydro_compiler_path = './../src/Hydro-Compiler'  # Path to Hydro-Compiler
        hydro_file = 'test_if.hy'  # The input .hy file
        output_executable = './../build/test_if'  # Name of the generated executable

        # Compile the .hy file to generate the executable
        compile_process = subprocess.run([hydro_compiler_path, hydro_file], capture_output=True)

        # Check if the compilation was successful
        self.assertEqual(
            compile_process.returncode, 0, 
            f"Hydro-Compiler failed with return code {compile_process.returncode}.\n"
            f"stderr: {compile_process.stderr.decode()}\n"
            f"stdout: {compile_process.stdout.decode()}"
        )

        # Run the generated executable
        if os.path.isfile(output_executable):
            run_process = subprocess.run([output_executable], capture_output=True)

            # Check the return code of the generated executable
            self.assertEqual(
                run_process.returncode, 2,
                f"Generated executable exited with code {run_process.returncode}, expected 2.\n"
                f"stderr: {run_process.stderr.decode()}\n"
                f"stdout: {run_process.stdout.decode()}"
            )
        else:
            self.fail(f"Expected output executable '{output_executable}' was not created.")
    def test_elif_statement(self):
        # Paths
        hydro_compiler_path = './../src/Hydro-Compiler'  # Path to Hydro-Compiler
        hydro_file = 'test_elif.hy'  # The input .hy file
        output_executable = './../build/test_elif'  # Name of the generated executable

        # Compile the .hy file to generate the executable
        compile_process = subprocess.run([hydro_compiler_path, hydro_file], capture_output=True)

        # Check if the compilation was successful
        self.assertEqual(
            compile_process.returncode, 0, 
            f"Hydro-Compiler failed with return code {compile_process.returncode}.\n"
            f"stderr: {compile_process.stderr.decode()}\n"
            f"stdout: {compile_process.stdout.decode()}"
        )

        # Run the generated executable
        if os.path.isfile(output_executable):
            run_process = subprocess.run([output_executable], capture_output=True)

            # Check the return code of the generated executable
            self.assertEqual(
                run_process.returncode, 1,
                f"Generated executable exited with code {run_process.returncode}, expected 1.\n"
                f"stderr: {run_process.stderr.decode()}\n"
                f"stdout: {run_process.stdout.decode()}"
            )
        else:
            self.fail(f"Expected output executable '{output_executable}' was not created.")
    def test_else_statement(self):
        # Paths
        hydro_compiler_path = './../src/Hydro-Compiler'  # Path to Hydro-Compiler
        hydro_file = 'test_else.hy'  # The input .hy file
        output_executable = './../build/test_else'  # Name of the generated executable

        # Compile the .hy file to generate the executable
        compile_process = subprocess.run([hydro_compiler_path, hydro_file], capture_output=True)

        # Check if the compilation was successful
        self.assertEqual(
            compile_process.returncode, 0, 
            f"Hydro-Compiler failed with return code {compile_process.returncode}.\n"
            f"stderr: {compile_process.stderr.decode()}\n"
            f"stdout: {compile_process.stdout.decode()}"
        )

        # Run the generated executable
        if os.path.isfile(output_executable):
            run_process = subprocess.run([output_executable], capture_output=True)

            # Check the return code of the generated executable
            self.assertEqual(
                run_process.returncode, 69,
                f"Generated executable exited with code {run_process.returncode}, expected 69.\n"
                f"stderr: {run_process.stderr.decode()}\n"
                f"stdout: {run_process.stdout.decode()}"
            )
        else:
            self.fail(f"Expected output executable '{output_executable}' was not created.")
    def test_binary_expressions(self):
        # Paths
        hydro_compiler_path = './../src/Hydro-Compiler'  # Path to Hydro-Compiler
        hydro_file = 'test_bin_expr.hy'  # The input .hy file
        output_executable = './../build/test_bin_expr'  # Name of the generated executable

        # Compile the .hy file to generate the executable
        compile_process = subprocess.run([hydro_compiler_path, hydro_file], capture_output=True)

        # Check if the compilation was successful
        self.assertEqual(
            compile_process.returncode, 0, 
            f"Hydro-Compiler failed with return code {compile_process.returncode}.\n"
            f"stderr: {compile_process.stderr.decode()}\n"
            f"stdout: {compile_process.stdout.decode()}"
        )

        # Run the generated executable
        if os.path.isfile(output_executable):
            run_process = subprocess.run([output_executable], capture_output=True)

            # Check the return code of the generated executable
            self.assertEqual(
                run_process.returncode, 12,
                f"Generated executable exited with code {run_process.returncode}, expected 12.\n"
                f"stderr: {run_process.stderr.decode()}\n"
                f"stdout: {run_process.stdout.decode()}"
            )
        else:
            self.fail(f"Expected output executable '{output_executable}' was not created.")
    def test_inner_scope(self):
        # Paths
        hydro_compiler_path = './../src/Hydro-Compiler'  # Path to Hydro-Compiler
        hydro_file = 'test_in_scope.hy'  # The input .hy file
        output_executable = './../build/test_in_scope'  # Name of the generated executable

        # Compile the .hy file to generate the executable
        compile_process = subprocess.run([hydro_compiler_path, hydro_file], capture_output=True)

        # Check if the compilation was successful
        self.assertEqual(
            compile_process.returncode, 0, 
            f"Hydro-Compiler failed with return code {compile_process.returncode}.\n"
            f"stderr: {compile_process.stderr.decode()}\n"
            f"stdout: {compile_process.stdout.decode()}"
        )

        # Run the generated executable
        if os.path.isfile(output_executable):
            run_process = subprocess.run([output_executable], capture_output=True)

            # Check the return code of the generated executable
            self.assertEqual(
                run_process.returncode, 3,
                f"Generated executable exited with code {run_process.returncode}, expected 3.\n"
                f"stderr: {run_process.stderr.decode()}\n"
                f"stdout: {run_process.stdout.decode()}"
            )
        else:
            self.fail(f"Expected output executable '{output_executable}' was not created.")
    def test_outer_scope(self):
        # Paths
        hydro_compiler_path = './../src/Hydro-Compiler'  # Path to Hydro-Compiler
        hydro_file = 'test_out_scope.hy'  # The input .hy file
        output_executable = './../build/test_out_scope'  # Name of the generated executable

        # Compile the .hy file to generate the executable
        compile_process = subprocess.run([hydro_compiler_path, hydro_file], capture_output=True)

        # Check if the compilation was successful
        self.assertEqual(
            compile_process.returncode, 0, 
            f"Hydro-Compiler failed with return code {compile_process.returncode}.\n"
            f"stderr: {compile_process.stderr.decode()}\n"
            f"stdout: {compile_process.stdout.decode()}"
        )

        # Run the generated executable
        if os.path.isfile(output_executable):
            run_process = subprocess.run([output_executable], capture_output=True)

            # Check the return code of the generated executable
            self.assertEqual(
                run_process.returncode, 1,
                f"Generated executable exited with code {run_process.returncode}, expected 1.\n"
                f"stderr: {run_process.stderr.decode()}\n"
                f"stdout: {run_process.stdout.decode()}"
            )
        else:
            self.fail(f"Expected output executable '{output_executable}' was not created.")





if __name__ == '__main__':
    unittest.main()
