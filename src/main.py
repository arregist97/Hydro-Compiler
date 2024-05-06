import sys

def main():
    # Check if arguments are provided
    if len(sys.argv) < 2:
        print("Usage: python main.py <argument>")
        return
    
    # Extract the argument(s)
    argument = sys.argv[1]
    
    # Do something with the argument
    print("Argument provided:", argument)

if __name__ == "__main__":
    main()
