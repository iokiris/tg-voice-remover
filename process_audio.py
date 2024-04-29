import sys
from spleeter.separator import Separator

def process_audio(input_audio, output_path):
    separator = Separator('spleeter:2stems')
    separator.separate_to_file(input_audio, output_path, codec='mp3')

if __name__ == '__main__':
    input_audio = sys.argv[1]
    output_path = sys.argv[2]
    process_audio(input_audio, output_path)
