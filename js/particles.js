particlesJS('particles-js', {
    particles: {
        number: {
            value: 80,
            density: {
                enable: true,
                value_area: 1500
            }
        },
        color: {
            value: ['#2563eb', '#60a5fa', '#93c5fd']
        },
        shape: {
            type: ['circle', 'triangle'],
            stroke: {
                width: 1,
                color: '#2563eb'
            }
        },
        opacity: {
            value: 0.1,
            random: true,
            anim: {
                enable: true,
                speed: 1,
                opacity_min: 0.05,
                sync: false
            }
        },
        size: {
            value: 3,
            random: true,
            anim: {
                enable: true,
                speed: 2,
                size_min: 0.5,
                sync: false
            }
        },
        line_linked: {
            enable: true,
            distance: 150,
            color: '#2563eb',
            opacity: 0.1,
            width: 1,
            shadow: {
                enable: true,
                blur: 5,
                color: '#2563eb'
            }
        },
        move: {
            enable: true,
            speed: 1,
            direction: 'none',
            random: true,
            straight: false,
            out_mode: 'out',
            bounce: false,
            attract: {
                enable: true,
                rotateX: 600,
                rotateY: 1200
            }
        }
    },
    interactivity: {
        detect_on: 'canvas',
        events: {
            onhover: {
                enable: true,
                mode: ['grab', 'bubble']
            },
            onclick: {
                enable: true,
                mode: 'push'
            },
            resize: true
        },
        modes: {
            grab: {
                distance: 140,
                line_linked: {
                    opacity: 0.4
                }
            },
            bubble: {
                distance: 200,
                size: 6,
                duration: 0.3,
                opacity: 0.8,
                speed: 3
            },
            push: {
                particles_nb: 3
            }
        }
    },
    retina_detect: true
});