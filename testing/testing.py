import subprocess
import unittest
import os
import shutil

class TestHydroCompiler(unittest.TestCase):
    def setUp(self):
        # Path to the build directory
        self.build_dir = './../build'
        self.hydro_compiler_path = './../src/Hydro-Compiler'

        # Clear the build directory
        if os.path.exists(self.build_dir):
            shutil.rmtree(self.build_dir)  # Remove all files and directories in the build directory
        os.makedirs(self.build_dir)  # Recreate the build directory

    def compile_and_run(self, hydro_file):
        """
        Compiles the given .hy file using the Hydro-Compiler and runs the resulting executable.

        Args:
            hydro_file (str): The name of the .hy file to compile.

        Returns:
            int: The return code of the generated executable.

        Raises:
            AssertionError: If the compilation fails or the executable is not created.
        """
        # Path for the generated executable
        output_executable = os.path.join(self.build_dir, os.path.splitext(hydro_file)[0])

        # Compile the .hy file to generate the executable
        compile_process = subprocess.run(
            [self.hydro_compiler_path, hydro_file],
            capture_output=True
        )

        # Ensure compilation was successful
        self.assertEqual(
            compile_process.returncode, 0,
            f"Hydro-Compiler failed with return code {compile_process.returncode}.\n"
            f"stderr: {compile_process.stderr.decode()}\n"
            f"stdout: {compile_process.stdout.decode()}"
        )

        # Ensure the executable was created
        self.assertTrue(
            os.path.isfile(output_executable),
            f"Expected output executable '{output_executable}' was not created."
        )

        # Run the generated executable
        run_process = subprocess.run([output_executable], capture_output=True)

        # Return the executable's return code
        return run_process.returncode

    def test_binary_expressions(self):
        return_code = self.compile_and_run('01_test_bin_expr.hy')
        self.assertEqual(
            return_code, 12,
            f"Executable for '01_test_bin_expr.hy' exited with code {return_code}, expected 12."
        )

    def test_inner_scope(self):
        return_code = self.compile_and_run('02_test_in_scope.hy')
        self.assertEqual(
            return_code, 3,
            f"Executable for '02_test_in_scope.hy' exited with code {return_code}, expected 3."
        )

    def test_outer_scope(self):
        return_code = self.compile_and_run('03_test_out_scope.hy')
        self.assertEqual(
            return_code, 1,
            f"Executable for '03_test_out_scope.hy' exited with code {return_code}, expected 1."
        )

    def test_if_statement(self):
        return_code = self.compile_and_run('04_test_if.hy')
        self.assertEqual(
            return_code, 2,
            f"Executable for '04_test_if.hy' exited with code {return_code}, expected 2."
        )

    def test_elif_statement(self):
        return_code = self.compile_and_run('05_test_elif.hy')
        self.assertEqual(
            return_code, 1,
            f"Executable for '05_test_elif.hy' exited with code {return_code}, expected 1."
        )
        
    def test_else_statement(self):
        return_code = self.compile_and_run('06_test_else.hy')
        self.assertEqual(
            return_code, 69,
            f"Executable for '06_test_else.hy' exited with code {return_code}, expected 69."
        )

    def test_multiline_statements(self):
        return_code = self.compile_and_run('07_test_mult_stmt.hy')
        self.assertEqual(
            return_code, 7,
            f"Executable for '07_test_mult_stmt.hy' exited with code {return_code}, expected 7."
        )


if __name__ == '__main__':
    unittest.main()
